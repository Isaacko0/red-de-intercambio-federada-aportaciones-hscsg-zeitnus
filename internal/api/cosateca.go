package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"federated-credit-node/internal/ledger"
)

// CosatecaHandler gestiona la cosateca/objetoteca del nodo: catalogo de objetos
// compartidos, prestamos, mantenimiento y aportaciones reconocidas en TQ.
//
// Mapeo al modelo Integral (devguide_v01):
//   OAD: cosateca_objects.design_ref -> diseno certificado en el repositorio OAD
//          (planos, lista de materiales, protocolo de mantenimiento).
//   COS: este modulo — organiza el acceso fisico a los objetos (prestamos,
//        mantenimiento, estado).
//   ITC/TQ: cosateca_contributions + ledger — la asamblea reconoce el aporte
//           de un objeto con creditos TQ (no transferibles, no dinero).
//   FRS: senales de uso/estado del objeto (retrasos, averias) para revision
//        del diseno en el OAD.
//
// Importante: TQ (este nodo) e ITC (Integral) son sistemas de contabilidad
// distintos; aqui se integra TQ como metrica de reciprocidad local, sin
// tratarlos como equivalentes.
type CosatecaHandler struct {
	Pool       *pgxpool.Pool
	NodeDomain string
}

func (h *CosatecaHandler) RegisterRoutes(r chi.Router, am *AuthMiddleware) {
	r.Group(func(r chi.Router) {
		r.Use(am.RequireAuth)
		r.Get("/api/cosateca/objects", h.listObjects)
		r.Post("/api/cosateca/objects", h.createObject)
		r.Get("/api/cosateca/objects/{id}", h.getObject)
		r.Patch("/api/cosateca/objects/{id}", h.updateObject)
		r.Delete("/api/cosateca/objects/{id}", h.deleteObject)

		r.Get("/api/cosateca/loans", h.listLoans)
		r.Post("/api/cosateca/loans", h.createLoan)
		r.Post("/api/cosateca/loans/{id}/return", h.returnLoan)

		r.Get("/api/cosateca/maintenance", h.listMaintenance)
		r.Post("/api/cosateca/maintenance", h.createMaintenance)

		r.Get("/api/cosateca/contributions", h.listContributions)
		r.Post("/api/cosateca/contributions", h.createContribution)
		r.Post("/api/cosateca/contributions/{id}/credit", h.creditContribution)
	})
}

// ---------------------------------------------------------------------------
// Objetos
// ---------------------------------------------------------------------------

type cosatecaObject struct {
	ID               string          `json:"id"`
	NodeDomain       string          `json:"node_domain"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	Category         string          `json:"category"`
	Condition        string          `json:"condition"`
	Status           string          `json:"status"`
	Location         string          `json:"location"`
	DesignRef        string          `json:"design_ref"`
	DesignVersion    string          `json:"design_version"`
	BillOfMaterials  json.RawMessage `json:"bill_of_materials"`
	TQValue          int64           `json:"tq_value"`
	AccessTQCost     int64           `json:"access_tq_cost"`
	ContributedBy    string          `json:"contributed_by"`
	ContributorName  string          `json:"contributor_name"`
	ContributedAt    time.Time       `json:"contributed_at"`
	FederatedVisible bool            `json:"federated_visible"`
	Images           json.RawMessage `json:"images"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	CurrentLoanID    *string         `json:"current_loan_id,omitempty"`
	CurrentBorrower  *string         `json:"current_borrower,omitempty"`
	LoanDueAt        *time.Time      `json:"loan_due_at,omitempty"`
	LoanState        *string         `json:"loan_state,omitempty"`
}

func (h *CosatecaHandler) listObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nodeDomain := h.NodeDomain

	status := r.URL.Query().Get("status")
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	query := `
		SELECT o.id, o.node_domain, o.name, o.description, o.category, o.condition, o.status,
		       o.location, o.design_ref, o.design_version, o.bill_of_materials, o.tq_value,
		       o.access_tq_cost, o.contributed_by, u.display_name, o.contributed_at,
		       o.federated_visible, o.images, o.created_at, o.updated_at,
		       l.id, l.borrower_id, l.due_at, l.state
		FROM cosateca_objects o
		LEFT JOIN users u ON u.id = o.contributed_by
		LEFT JOIN LATERAL (
			SELECT id, borrower_id, due_at, state
			FROM cosateca_loans
			WHERE object_id = o.id AND state = 'activo'
			ORDER BY loaned_at DESC LIMIT 1
		) l ON true
		WHERE o.node_domain = $1`
	args := []interface{}{nodeDomain}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND o.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if category != "" {
		query += fmt.Sprintf(" AND o.category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}
	if search != "" {
		query += fmt.Sprintf(" AND (o.name ILIKE $%d OR o.description ILIKE $%d)", argIdx, argIdx+1)
		args = append(args, "%"+search+"%", "%"+search+"%")
		argIdx += 2
	}
	query += fmt.Sprintf(" ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := h.Pool.Query(ctx, query, args...)
	if err != nil {
		writeError(w, 500, "error listing objects")
		return
	}
	defer rows.Close()

	objects := []cosatecaObject{}
	for rows.Next() {
		var o cosatecaObject
		var bmRaw, imagesRaw string
		var loanID, borrowerID *string
		var dueAt *time.Time
		var loanState *string
		if err := rows.Scan(&o.ID, &o.NodeDomain, &o.Name, &o.Description, &o.Category,
			&o.Condition, &o.Status, &o.Location, &o.DesignRef, &o.DesignVersion,
			&bmRaw, &o.TQValue, &o.AccessTQCost, &o.ContributedBy, &o.ContributorName,
			&o.ContributedAt, &o.FederatedVisible, &imagesRaw, &o.CreatedAt, &o.UpdatedAt,
			&loanID, &borrowerID, &dueAt, &loanState); err != nil {
			writeError(w, 500, "error scanning object")
			return
		}
		if bmRaw != "" {
			o.BillOfMaterials = json.RawMessage(bmRaw)
		}
		if imagesRaw != "" {
			o.Images = json.RawMessage(imagesRaw)
		}
		o.CurrentLoanID = loanID
		o.CurrentBorrower = borrowerID
		o.LoanDueAt = dueAt
		o.LoanState = loanState
		objects = append(objects, o)
	}
	writeJSON(w, 200, map[string]interface{}{"objects": objects, "count": len(objects)})
}

type createObjectRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	Condition       string   `json:"condition"`
	Status          string   `json:"status"`
	Location        string   `json:"location"`
	DesignRef       string   `json:"design_ref"`
	DesignVersion   string   `json:"design_version"`
	BillOfMaterials []string `json:"bill_of_materials"`
	TQValue         int64    `json:"tq_value"`
	AccessTQCost    int64    `json:"access_tq_cost"`
	FederatedVisible bool    `json:"federated_visible"`
	Images          []string `json:"images"`
}

func (h *CosatecaHandler) createObject(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, 401, "usuario no identificado")
		return
	}

	var req createObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "cuerpo de solicitud invalido")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, 400, "el nombre es obligatorio")
		return
	}
	if req.Category == "" {
		req.Category = "herramienta"
	}
	if req.Condition == "" {
		req.Condition = "bueno"
	}
	if req.Status == "" {
		req.Status = "disponible"
	}

	bm, _ := json.Marshal(req.BillOfMaterials)
	imgs, _ := json.Marshal(req.Images)

	var bmJSON pgtype.JSONB
	_ = bmJSON.Set(bm)
	var imgsJSON pgtype.JSONB
	_ = imgsJSON.Set(imgs)

	var id uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO cosateca_objects
			(node_domain, name, description, category, condition, status, location,
			 design_ref, design_version, bill_of_materials, tq_value, access_tq_cost,
			 contributed_by, federated_visible, images)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`,
		h.NodeDomain, req.Name, req.Description, req.Category, req.Condition, req.Status,
		req.Location, req.DesignRef, req.DesignVersion, bmJSON, req.TQValue,
		req.AccessTQCost, userID, req.FederatedVisible, imgsJSON).Scan(&id)
	if err != nil {
		writeError(w, 500, "error creando objeto")
		return
	}
	writeJSON(w, 201, map[string]interface{}{"id": id.String(), "status": "created"})
}

func (h *CosatecaHandler) getObject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, 400, "id invalido")
		return
	}
	// reutiliza la lista con filtro por id
	query := `
		SELECT o.id, o.node_domain, o.name, o.description, o.category, o.condition, o.status,
		       o.location, o.design_ref, o.design_version, o.bill_of_materials, o.tq_value,
		       o.access_tq_cost, o.contributed_by, u.display_name, o.contributed_at,
		       o.federated_visible, o.images, o.created_at, o.updated_at,
		       l.id, l.borrower_id, l.due_at, l.state
		FROM cosateca_objects o
		LEFT JOIN users u ON u.id = o.contributed_by
		LEFT JOIN LATERAL (
			SELECT id, borrower_id, due_at, state
			FROM cosateca_loans
			WHERE object_id = o.id AND state = 'activo'
			ORDER BY loaned_at DESC LIMIT 1
		) l ON true
		WHERE o.id = $1 AND o.node_domain = $2`
	var o cosatecaObject
	var bmRaw, imagesRaw string
	var loanID, borrowerID *string
	var dueAt *time.Time
	var loanState *string
	err = h.Pool.QueryRow(r.Context(), query, id, h.NodeDomain).Scan(
		&o.ID, &o.NodeDomain, &o.Name, &o.Description, &o.Category, &o.Condition, &o.Status,
		&o.Location, &o.DesignRef, &o.DesignVersion, &bmRaw, &o.TQValue, &o.AccessTQCost,
		&o.ContributedBy, &o.ContributorName, &o.ContributedAt, &o.FederatedVisible,
		&imagesRaw, &o.CreatedAt, &o.UpdatedAt, &loanID, &borrowerID, &dueAt, &loanState)
	if err == pgx.ErrNoRows {
		writeError(w, 404, "objeto no encontrado")
		return
	}
	if err != nil {
		writeError(w, 500, "error obteniendo objeto")
		return
	}
	if bmRaw != "" {
		o.BillOfMaterials = []byte(bmRaw)
	}
	if imagesRaw != "" {
		o.Images = []byte(imagesRaw)
	}
	o.CurrentLoanID = loanID
	o.CurrentBorrower = borrowerID
	o.LoanDueAt = dueAt
	o.LoanState = loanState
	writeJSON(w, 200, o)
}

type updateObjectRequest struct {
	Name            *string  `json:"name"`
	Description     *string  `json:"description"`
	Category        *string  `json:"category"`
	Condition       *string  `json:"condition"`
	Status          *string  `json:"status"`
	Location        *string  `json:"location"`
	DesignRef       *string  `json:"design_ref"`
	DesignVersion   *string  `json:"design_version"`
	TQValue         *int64   `json:"tq_value"`
	AccessTQCost    *int64   `json:"access_tq_cost"`
	FederatedVisible *bool   `json:"federated_visible"`
}

func (h *CosatecaHandler) updateObject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, 400, "id invalido")
		return
	}
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	var contributorID string
	err = h.Pool.QueryRow(r.Context(), `SELECT contributed_by FROM cosateca_objects WHERE id = $1 AND node_domain = $2`,
		id, h.NodeDomain).Scan(&contributorID)
	if err == pgx.ErrNoRows {
		writeError(w, 404, "objeto no encontrado")
		return
	}
	if err != nil {
		writeError(w, 500, "error verificando objeto")
		return
	}
	if contributorID != userID.String() {
		writeError(w, 403, "solo quien aportó el objeto puede editarlo")
		return
	}

	var req updateObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "cuerpo de solicitud invalido")
		return
	}

	sets := []string{}
	args := []interface{}{id, h.NodeDomain}
	argIdx := 3
	setField := func(col, val string) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}
	if req.Name != nil {
		setField("name", *req.Name)
	}
	if req.Description != nil {
		setField("description", *req.Description)
	}
	if req.Category != nil {
		setField("category", *req.Category)
	}
	if req.Condition != nil {
		setField("condition", *req.Condition)
	}
	if req.Status != nil {
		setField("status", *req.Status)
	}
	if req.Location != nil {
		setField("location", *req.Location)
	}
	if req.DesignRef != nil {
		setField("design_ref", *req.DesignRef)
	}
	if req.DesignVersion != nil {
		setField("design_version", *req.DesignVersion)
	}
	if req.TQValue != nil {
		setField("tq_value", fmt.Sprintf("%d", *req.TQValue))
	}
	if req.AccessTQCost != nil {
		setField("access_tq_cost", fmt.Sprintf("%d", *req.AccessTQCost))
	}
	if req.FederatedVisible != nil {
		setField("federated_visible", fmt.Sprintf("%v", *req.FederatedVisible))
	}
	if len(sets) == 0 {
		writeError(w, 400, "no hay campos para actualizar")
		return
	}
	sets = append(sets, fmt.Sprintf("updated_at = NOW()"))
	query := "UPDATE cosateca_objects SET " + strings.Join(sets, ", ") +
		fmt.Sprintf(" WHERE id = $1 AND node_domain = $2")
	if _, err := h.Pool.Exec(r.Context(), query, args...); err != nil {
		writeError(w, 500, "error actualizando objeto")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"id": id.String(), "status": "updated"})
}

func (h *CosatecaHandler) deleteObject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, 400, "id invalido")
		return
	}
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	var contributorID string
	var activeLoans int
	err = h.Pool.QueryRow(r.Context(), `
		SELECT contributed_by,
		       (SELECT COUNT(*) FROM cosateca_loans WHERE object_id = $1 AND state = 'activo')
		FROM cosateca_objects WHERE id = $2 AND node_domain = $3`,
		id, id, h.NodeDomain).Scan(&contributorID, &activeLoans)
	if err == pgx.ErrNoRows {
		writeError(w, 404, "objeto no encontrado")
		return
	}
	if err != nil {
		writeError(w, 500, "error verificando objeto")
		return
	}
	if contributorID != userID.String() {
		writeError(w, 403, "solo quien aportó el objeto puede eliminarlo")
		return
	}
	if activeLoans > 0 {
		writeError(w, 400, "no se puede eliminar un objeto con prestamos activos")
		return
	}
	if _, err := h.Pool.Exec(r.Context(), "DELETE FROM cosateca_objects WHERE id = $1 AND node_domain = $2", id, h.NodeDomain); err != nil {
		writeError(w, 500, "error eliminando objeto")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"id": id.String(), "status": "deleted"})
}

// ---------------------------------------------------------------------------
// Prestamos
// ---------------------------------------------------------------------------

type cosatecaLoan struct {
	ID          string     `json:"id"`
	ObjectID    string     `json:"object_id"`
	ObjectName  string     `json:"object_name"`
	BorrowerID  string     `json:"borrower_id"`
	BorrowerName string    `json:"borrower_name"`
	LoanedAt    time.Time  `json:"loaned_at"`
	DueAt       *time.Time `json:"due_at"`
	ReturnedAt  *time.Time `json:"returned_at"`
	State       string     `json:"state"`
	Notes       string     `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
}

type createLoanRequest struct {
	ObjectID string     `json:"object_id"`
	DueAt    *time.Time `json:"due_at"`
	Notes    string     `json:"notes"`
}

func (h *CosatecaHandler) listLoans(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)
	mine := r.URL.Query().Get("mine") == "1"
	state := r.URL.Query().Get("state")

	query := `
		SELECT l.id, l.object_id, o.name, l.borrower_id, u.display_name, l.loaned_at,
		       l.due_at, l.returned_at, l.state, l.notes, l.created_at
		FROM cosateca_loans l
		JOIN cosateca_objects o ON o.id = l.object_id
		JOIN users u ON u.id = l.borrower_id
		WHERE o.node_domain = $1`
	args := []interface{}{h.NodeDomain}
	argIdx := 2
	if mine {
		query += fmt.Sprintf(" AND l.borrower_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	}
	if state != "" {
		query += fmt.Sprintf(" AND l.state = $%d", argIdx)
		args = append(args, state)
		argIdx++
	}
	query += " ORDER BY l.created_at DESC"

	rows, err := h.Pool.Query(r.Context(), query, args...)
	if err != nil {
		writeError(w, 500, "error listando prestamos")
		return
	}
	defer rows.Close()

	loans := []cosatecaLoan{}
	for rows.Next() {
		var l cosatecaLoan
		var dueAt, returnedAt *time.Time
		if err := rows.Scan(&l.ID, &l.ObjectID, &l.ObjectName, &l.BorrowerID, &l.BorrowerName,
			&l.LoanedAt, &dueAt, &returnedAt, &l.State, &l.Notes, &l.CreatedAt); err != nil {
			writeError(w, 500, "error escaneando prestamo")
			return
		}
		l.DueAt = dueAt
		l.ReturnedAt = returnedAt
		loans = append(loans, l)
	}
	writeJSON(w, 200, map[string]interface{}{"loans": loans, "count": len(loans)})
}

func (h *CosatecaHandler) createLoan(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, 401, "usuario no identificado")
		return
	}

	var req createLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "cuerpo de solicitud invalido")
		return
	}
	objectID, err := uuid.Parse(req.ObjectID)
	if err != nil {
		writeError(w, 400, "id de objeto invalido")
		return
	}

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		writeError(w, 500, "error iniciando transaccion")
		return
	}
	defer tx.Rollback(r.Context())

	var status, contributorID string
	err = tx.QueryRow(r.Context(), `
		SELECT status, contributed_by FROM cosateca_objects WHERE id = $1 AND node_domain = $2 FOR UPDATE`,
		objectID, h.NodeDomain).Scan(&status, &contributorID)
	if err == pgx.ErrNoRows {
		writeError(w, 404, "objeto no encontrado")
		return
	}
	if err != nil {
		writeError(w, 500, "error verificando objeto")
		return
	}
	if status != "disponible" {
		writeError(w, 409, "el objeto no esta disponible para prestamo")
		return
	}
	if contributorID == userID.String() {
		writeError(w, 400, "no puedes prestarte tu propio objeto aportado")
		return
	}

	var loanID uuid.UUID
	err = tx.QueryRow(r.Context(), `
		INSERT INTO cosateca_loans (object_id, borrower_id, due_at, state)
		VALUES ($1, $2, $3, 'activo') RETURNING id`,
		objectID, userID, req.DueAt).Scan(&loanID)
	if err != nil {
		writeError(w, 500, "error creando prestamo")
		return
	}
	if _, err := tx.Exec(r.Context(), "UPDATE cosateca_objects SET status = 'prestado', updated_at = NOW() WHERE id = $1", objectID); err != nil {
		writeError(w, 500, "error actualizando estado del objeto")
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, 500, "error confirmando prestamo")
		return
	}
	writeJSON(w, 201, map[string]interface{}{"id": loanID.String(), "status": "created"})
}

func (h *CosatecaHandler) returnLoan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	loanID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, 400, "id invalido")
		return
	}
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		writeError(w, 500, "error iniciando transaccion")
		return
	}
	defer tx.Rollback(r.Context())

	var objectID uuid.UUID
	var borrowerID string
	var state string
	err = tx.QueryRow(r.Context(), `
		SELECT object_id, borrower_id, state FROM cosateca_loans WHERE id = $1`,
		loanID).Scan(&objectID, &borrowerID, &state)
	if err == pgx.ErrNoRows {
		writeError(w, 404, "prestamo no encontrado")
		return
	}
	if err != nil {
		writeError(w, 500, "error verificando prestamo")
		return
	}
	if state != "activo" {
		writeError(w, 409, "el prestamo ya no esta activo")
		return
	}
	if borrowerID != userID.String() {
		writeError(w, 403, "solo quien pidio prestado el objeto puede devolverlo")
		return
	}

	if _, err := tx.Exec(r.Context(), `
		UPDATE cosateca_loans SET state = 'devuelto', returned_at = NOW() WHERE id = $1`, loanID); err != nil {
		writeError(w, 500, "error devolviendo prestamo")
		return
	}
	if _, err := tx.Exec(r.Context(), `
		UPDATE cosateca_objects SET status = 'disponible', updated_at = NOW() WHERE id = $1`, objectID); err != nil {
		writeError(w, 500, "error actualizando objeto")
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, 500, "error confirmando devolucion")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"id": idStr, "status": "returned"})
}

// ---------------------------------------------------------------------------
// Mantenimiento
// ---------------------------------------------------------------------------

type cosatecaMaintenance struct {
	ID               string     `json:"id"`
	ObjectID         string     `json:"object_id"`
	ObjectName       string     `json:"object_name"`
	PerformedBy      string     `json:"performed_by"`
	PerformedByName  string     `json:"performed_by_name"`
	PerformedAt      time.Time  `json:"performed_at"`
	Type             string     `json:"type"`
	Description      string     `json:"description"`
	PartsCostTQ      int64      `json:"parts_cost_tq"`
	Outcome          string     `json:"outcome"`
	NextMaintenance  *time.Time `json:"next_maintenance_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

type createMaintenanceRequest struct {
	ObjectID         string     `json:"object_id"`
	Type             string     `json:"type"`
	Description      string     `json:"description"`
	PartsCostTQ      int64      `json:"parts_cost_tq"`
	Outcome          string     `json:"outcome"`
	NextMaintenance  *time.Time `json:"next_maintenance_at"`
}

func (h *CosatecaHandler) listMaintenance(w http.ResponseWriter, r *http.Request) {
	objectID := r.URL.Query().Get("object_id")
	query := `
		SELECT m.id, m.object_id, o.name, m.performed_by, u.display_name, m.performed_at,
		       m.type, m.description, m.parts_cost_tq, m.outcome, m.next_maintenance_at,
		       m.created_at
		FROM cosateca_maintenance m
		JOIN cosateca_objects o ON o.id = m.object_id
		JOIN users u ON u.id = m.performed_by
		WHERE o.node_domain = $1`
	args := []interface{}{h.NodeDomain}
	if objectID != "" {
		query += " AND m.object_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, objectID)
	}
	query += " ORDER BY m.performed_at DESC"

	rows, err := h.Pool.Query(r.Context(), query, args...)
	if err != nil {
		writeError(w, 500, "error listando mantenimiento")
		return
	}
	defer rows.Close()

	items := []cosatecaMaintenance{}
	for rows.Next() {
		var m cosatecaMaintenance
		var next *time.Time
		if err := rows.Scan(&m.ID, &m.ObjectID, &m.ObjectName, &m.PerformedBy,
			&m.PerformedByName, &m.PerformedAt, &m.Type, &m.Description,
			&m.PartsCostTQ, &m.Outcome, &next, &m.CreatedAt); err != nil {
			writeError(w, 500, "error escaneando mantenimiento")
			return
		}
		m.NextMaintenance = next
		items = append(items, m)
	}
	writeJSON(w, 200, map[string]interface{}{"maintenance": items, "count": len(items)})
}

func (h *CosatecaHandler) createMaintenance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, 401, "usuario no identificado")
		return
	}

	var req createMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "cuerpo de solicitud invalido")
		return
	}
	if req.Type == "" {
		writeError(w, 400, "el tipo de mantenimiento es obligatorio")
		return
	}
	objectID, err := uuid.Parse(req.ObjectID)
	if err != nil {
		writeError(w, 400, "id de objeto invalido")
		return
	}

	var exists bool
	if err := h.Pool.QueryRow(r.Context(), `
		SELECT EXISTS(SELECT 1 FROM cosateca_objects WHERE id = $1 AND node_domain = $2)`,
		objectID, h.NodeDomain).Scan(&exists); err != nil || !exists {
		writeError(w, 404, "objeto no encontrado")
		return
	}

	var mID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO cosateca_maintenance (object_id, performed_by, type, description,
		                                  parts_cost_tq, outcome, next_maintenance_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		objectID, userID, req.Type, req.Description, req.PartsCostTQ,
		req.Outcome, req.NextMaintenance).Scan(&mID)
	if err != nil {
		writeError(w, 500, "error registrando mantenimiento")
		return
	}
	// Si se marca reparacion, el objeto vuelve a disponible tras la revision
	if req.Type == "reparacion" && req.Outcome == "ok" {
		_, _ = h.Pool.Exec(r.Context(), `
			UPDATE cosateca_objects SET status = 'disponible', updated_at = NOW() WHERE id = $1`, objectID)
	}
	writeJSON(w, 201, map[string]interface{}{"id": mID.String(), "status": "created"})
}

// ---------------------------------------------------------------------------
// Aportaciones (reciprocidad en TQ)
// ---------------------------------------------------------------------------

type cosatecaContribution struct {
	ID              string     `json:"id"`
	ObjectID        string     `json:"object_id"`
	ObjectName      string     `json:"object_name"`
	ContributorID   string     `json:"contributor_id"`
	ContributorName string     `json:"contributor_name"`
	TQValue         int64      `json:"tq_value"`
	Credited        bool       `json:"credited"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type createContributionRequest struct {
	ObjectID string `json:"object_id"`
	TQValue  int64  `json:"tq_value"`
}

func (h *CosatecaHandler) listContributions(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT c.id, c.object_id, o.name, c.contributor_id, u.display_name, c.tq_value,
		       c.credited, c.acknowledged_at, c.created_at
		FROM cosateca_contributions c
		JOIN cosateca_objects o ON o.id = c.object_id
		JOIN users u ON u.id = c.contributor_id
		WHERE o.node_domain = $1
		ORDER BY c.created_at DESC`
	rows, err := h.Pool.Query(r.Context(), query, h.NodeDomain)
	if err != nil {
		writeError(w, 500, "error listando aportaciones")
		return
	}
	defer rows.Close()

	items := []cosatecaContribution{}
	for rows.Next() {
		var c cosatecaContribution
		var ack *time.Time
		if err := rows.Scan(&c.ID, &c.ObjectID, &c.ObjectName, &c.ContributorID,
			&c.ContributorName, &c.TQValue, &c.Credited, &ack, &c.CreatedAt); err != nil {
			writeError(w, 500, "error escaneando aportacion")
			return
		}
		c.AcknowledgedAt = ack
		items = append(items, c)
	}
	writeJSON(w, 200, map[string]interface{}{"contributions": items, "count": len(items)})
}

func (h *CosatecaHandler) createContribution(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, 401, "usuario no identificado")
		return
	}

	var req createContributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "cuerpo de solicitud invalido")
		return
	}
	objectID, err := uuid.Parse(req.ObjectID)
	if err != nil {
		writeError(w, 400, "id de objeto invalido")
		return
	}

	var contributorID string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT contributed_by FROM cosateca_objects WHERE id = $1 AND node_domain = $2`,
		objectID, h.NodeDomain).Scan(&contributorID)
	if err == pgx.ErrNoRows {
		writeError(w, 404, "objeto no encontrado")
		return
	}
	if err != nil {
		writeError(w, 500, "error verificando objeto")
		return
	}
	if contributorID != userID.String() {
		writeError(w, 403, "solo quien aportó el objeto puede registrar la aportacion")
		return
	}

	var cID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO cosateca_contributions (object_id, contributor_id, tq_value)
		VALUES ($1, $2, $3) RETURNING id`,
		objectID, userID, req.TQValue).Scan(&cID)
	if err != nil {
		writeError(w, 500, "error registrando aportacion")
		return
	}
	writeJSON(w, 201, map[string]interface{}{"id": cID.String(), "status": "created"})
}

// creditContribution acredita el aporte en el ledger TQ del nodo:
// credito al balance del contribuyente (user_balance), contrapartida al puente
// nodal (node_bridge_global), siguiendo el mismo patron de escritura directa que
// usa internal/federation para asientos de puente. No es una transferencia entre
// miembros: es la comunidad reconociendo una aportacion al fondo colectivo.
func (h *CosatecaHandler) creditContribution(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	cID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, 400, "id invalido")
		return
	}
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	var contributorID string
	var tqValue int64
	var credited bool
	err = h.Pool.QueryRow(r.Context(), `
		SELECT contributor_id, tq_value, credited FROM cosateca_contributions WHERE id = $1`,
		cID).Scan(&contributorID, &tqValue, &credited)
	if err == pgx.ErrNoRows {
		writeError(w, 404, "aportacion no encontrada")
		return
	}
	if err != nil {
		writeError(w, 500, "error verificando aportacion")
		return
	}
	if credited {
		writeError(w, 409, "la aportacion ya fue acreditada en el ledger TQ")
		return
	}
	if contributorID != userID.String() {
		writeError(w, 403, "solo quien hizo la aportacion puede solicitar el credito")
		return
	}
	if tqValue <= 0 {
		writeError(w, 400, "la aportacion no tiene valor TQ reconocido para acreditar")
		return
	}

	nilUUID := uuid.UUID{}

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		writeError(w, 500, "error iniciando transaccion")
		return
	}
	defer tx.Rollback(r.Context())

	var userNodeDomain string
	if err := tx.QueryRow(r.Context(), `SELECT node_domain FROM users WHERE id = $1`, userID).Scan(&userNodeDomain); err != nil {
		writeError(w, 404, "usuario no encontrado")
		return
	}

	metadata := map[string]interface{}{
		"type":            "cosateca_contribution",
		"contribution_id": cID.String(),
		"contributor_id":  contributorID,
		"object_domain":   h.NodeDomain,
	}

	var txID uuid.UUID
	err = tx.QueryRow(r.Context(), `
		INSERT INTO transactions (id, tx_type, sender_id, receiver_id, amount, status, metadata, created_at)
		VALUES ($1, $2, NULL, $3, $4, 'confirmed', $5, NOW())
		RETURNING id`,
		uuid.New(), ledger.TxTypeCosatecaContribution, userID, tqValue, metadataJSON).Scan(&txID)
	if err != nil {
		writeError(w, 500, "error creando transaccion TQ")
		return
	}

	nilUUIDStr := nilUUID.String()
	_, _ = tx.Exec(r.Context(), `
		INSERT INTO ledger_entries (transaction_id, account_id, entry_type, amount, account_category, counterpart_node, created_at)
		VALUES ($1, $2, 'credit', $3, 'user_balance', NULL, NOW()),
		       ($1, $4, 'debit', $3, 'node_bridge_global', $5, NOW())`,
		txID, userID, tqValue, nilUUIDStr, h.NodeDomain)

	if _, err := tx.Exec(r.Context(), `
		UPDATE cosateca_contributions SET credited = true, acknowledged_at = NOW() WHERE id = $1`, cID); err != nil {
		writeError(w, 500, "error marcando aportacion como acreditada")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, 500, "error confirmando acreditacion")
		return
	}
	writeJSON(w, 200, map[string]interface{}{
		"id":          cID.String(),
		"status":      "credited",
		"tq_value":    tqValue,
		"contributor": contributorID,
		"node_domain": h.NodeDomain,
	})
}
