-- Cosateca / Objetoteca: biblioteca comunitaria de objetos y herramientas compartidas.
--
-- Contexto: ciclo OAD -> COS -> ITC -> FRS de Integral (Diseño de Acceso Abierto,
-- Organisacion Cooperativa, Creditos de Contribucion, Feedback y Revision) aplicado
-- a la red de ecoaldeas federadas, con la metrica TQ del Libro Ecoaldeas Federadas
-- (1 TQ = 1 kWh, cinta metrica de reciprocidad, no dinero acumulable).
--
-- Principio disenado: el OBJETO (fisico), el REGISTRO (disponibilidad, prestamos,
-- mantenimiento) y la RECIPROCIDAD (aportaciones reconocidas en TQ por la asamblea)
-- son tres cosas distintas. Nunca se confunden.
--
-- Integracion con OAD: cosateca_objects.design_ref referencia un diseno certificado
-- en el repositorio OAD (planos, lista de materiales, protocolo de mantenimiento).
-- Un objeto solo se pone en produccion/prestamo si el diseno de referencia existe
-- en el OAD (igual que COS no organiza trabajo sin diseno certificado).

CREATE TABLE IF NOT EXISTS cosateca_objects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  node_domain TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  category TEXT NOT NULL DEFAULT 'herramienta',
  condition TEXT NOT NULL DEFAULT 'bueno', -- bueno, regular, malo, fuera_de_servicio
  status TEXT NOT NULL DEFAULT 'disponible', -- disponible, prestado, mantenimiento, retirado
  location TEXT,
  design_ref TEXT, -- referencia al diseno certificado en el OAD (Open Access Design)
  design_version TEXT,
  bill_of_materials JSONB DEFAULT '[]'::jsonb,
  tq_value BIGINT NOT NULL DEFAULT 0, -- energia incorporada del objeto en TQ (1 TQ = 1 kWh)
  access_tq_cost BIGINT NOT NULL DEFAULT 0, -- costo de acceso en TQ si la asamblea lo decide
  contributed_by UUID NOT NULL REFERENCES users(id),
  contributed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  federated_visible BOOLEAN NOT NULL DEFAULT false,
  images JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cosateca_objects_node_domain ON cosateca_objects(node_domain);
CREATE INDEX IF NOT EXISTS idx_cosateca_objects_status ON cosateca_objects(status);
CREATE INDEX IF NOT EXISTS idx_cosateca_objects_category ON cosateca_objects(category);
CREATE INDEX IF NOT EXISTS idx_cosateca_objects_design_ref ON cosateca_objects(design_ref);
CREATE INDEX IF NOT EXISTS idx_cosateca_objects_contributed_by ON cosateca_objects(contributed_by);

CREATE TABLE IF NOT EXISTS cosateca_loans (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  object_id UUID NOT NULL REFERENCES cosateca_objects(id) ON DELETE CASCADE,
  borrower_id UUID NOT NULL REFERENCES users(id),
  loaned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  due_at TIMESTAMPTZ,
  returned_at TIMESTAMPTZ,
  state TEXT NOT NULL DEFAULT 'activo', -- activo, devuelto, vencido
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cosateca_loans_object ON cosateca_loans(object_id);
CREATE INDEX IF NOT EXISTS idx_cosateca_loans_borrower ON cosateca_loans(borrower_id);
CREATE INDEX IF NOT EXISTS idx_cosateca_loans_state ON cosateca_loans(state);

CREATE TABLE IF NOT EXISTS cosateca_maintenance (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  object_id UUID NOT NULL REFERENCES cosateca_objects(id) ON DELETE CASCADE,
  performed_by UUID NOT NULL REFERENCES users(id),
  performed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  type TEXT NOT NULL, -- reparacion, limpieza, calibracion, revision
  description TEXT,
  parts_cost_tq BIGINT NOT NULL DEFAULT 0,
  outcome TEXT,
  next_maintenance_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cosateca_maintenance_object ON cosateca_maintenance(object_id);
CREATE INDEX IF NOT EXISTS idx_cosateca_maintenance_performed_by ON cosateca_maintenance(performed_by);

CREATE TABLE IF NOT EXISTS cosateca_contributions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  object_id UUID NOT NULL REFERENCES cosateca_objects(id) ON DELETE CASCADE,
  contributor_id UUID NOT NULL REFERENCES users(id),
  tq_value BIGINT NOT NULL DEFAULT 0, -- valor reconocido por la asamblea en TQ
  credited BOOLEAN NOT NULL DEFAULT false, -- si ya se acreditó en el ledger TQ
  acknowledged_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cosateca_contributions_object ON cosateca_contributions(object_id);
CREATE INDEX IF NOT EXISTS idx_cosateca_contributions_contributor ON cosateca_contributions(contributor_id);
CREATE INDEX IF NOT EXISTS idx_cosateca_contributions_credited ON cosateca_contributions(credited);

COMMENT ON TABLE cosateca_objects IS 'Catalogo de objetos de la cosateca/objetoteca del nodo (registro de bienes comunes fisicos).';
COMMENT ON TABLE cosateca_loans IS 'Historial de prestamos de objetos de la cosateca.';
COMMENT ON TABLE cosateca_maintenance IS 'Mantenimiento y reparaciones de objetos de la cosateca.';
COMMENT ON TABLE cosateca_contributions IS 'Aportaciones de objetos al fondo colectivo, reconocidas en TQ por la asamblea del nodo.';
