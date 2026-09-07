#!/usr/bin/env node
/**
 * RIDF Orchestrator - script de orquestación para red-de-intercambio-federada-isaacko
 * 
 * Comandos:
 *   status        - Muestra estado actual + menú interactivo
 *   graph         - Grafo de dependencias
 *   next          - Próxima tarea recomendada
 *   done <id>     - Marcar tarea como completada
 *   add-task      - Añadir tarea custom
 *   init          - Inicializar con tareas base
 */

const fs = require('fs');
const path = require('path');

const STATE_FILE = path.join(__dirname, '..', 'ridf-orchestrator-state.json');

// Colores
const RESET = '\x1b[0m';
const BOLD = '\x1b[1m';
const GREEN = '\x1b[32m';
const YELLOW = '\x1b[33m';
const RED = '\x1b[31m';
const CYAN = '\x1b[36m';
const GRAY = '\x1b[90m';

// Cargar estado
function loadState() {
    if (!fs.existsSync(STATE_FILE)) {
        console.log(`${YELLOW}⚠️  No se encontró ${STATE_FILE}`);
        console.log(`   Ejecuta: node scripts/orchestrator-next-steps.js init${RESET}`);
        process.exit(1);
    }
    return JSON.parse(fs.readFileSync(STATE_FILE, 'utf8'));
}

// Guardar estado
function saveState(state) {
    state.lastUpdated = new Date().toISOString();
    fs.writeFileSync(STATE_FILE, JSON.stringify(state, null, 2));
    console.log(`${GRAY}Estado guardado en ${STATE_FILE}${RESET}`);
}

// Calcular score de priorización
function calculateScore(task, allTasks) {
    if (task.status !== 'pending') return -1;
    
    // Verificar dependencias
    const depsDone = (task.dependencies || []).every(depId => {
        const dep = allTasks.find(t => t.id === depId);
        return dep && dep.status === 'done';
    });
    if (!depsDone) return -1;
    
    // Contar desbloqueos
    const unblocks = allTasks.filter(t => 
        (t.dependencies || []).includes(task.id) && t.status === 'pending'
    ).length;
    
    const score = (task.value * 0.5) + ((6 - task.effort) * 10 * 0.3) + (unblocks * 15 * 0.2);
    return Math.round(score * 10) / 10;
}

// Mostrar estado (menú interactivo)
function showStatus(state) {
    console.log(`\n${BOLD}${CYAN}═══════════════════════════════════════════════════════${RESET}`);
    console.log(`${BOLD}${CYAN}  🤖 RIDF ORCHESTRATOR — Motor de Inversión PSG${RESET}`);
    console.log(`${BOLD}${CYAN}═════════════════════════════════════════════════════════════${RESET}\n`);
    
    // Agrupar por workstream
    const workstreams = {};
    for (const task of state.taskRegistry) {
        if (!workstreams[task.workstream]) {
            workstreams[task.workstream] = [];
        }
        workstreams[task.workstream].push(task);
    }
    
    for (const [ws, tasks] of Object.entries(workstreams)) {
        console.log(`${BOLD}${YELLOW}📂 ${ws}${RESET}`);
        for (const task of tasks) {
            const statusIcon = task.status === 'done' ? '✅' : 
                              task.status === 'in_progress' ? '🔄' :
                              task.status === 'blocked' ? '🚫' : '⬜';
            const score = calculateScore(task, state.taskRegistry);
            const scoreStr = score >= 0 ? `${GRAY}score: ${score}${RESET}` : '';
            console.log(`   ${statusIcon} ${task.id} — ${task.title} ${scoreStr}`);
        }
        console.log();
    }
    
    // Próxima tarea recomendada
    const available = state.taskRegistry
        .map(t => ({ ...t, score: calculateScore(t, state.taskRegistry) }))
        .filter(t => t.score >= 0)
        .sort((a, b) => b.score - a.score);
    
    if (available.length > 0) {
        const next = available[0];
        console.log(`${BOLD}${GREEN}🎯 Próxima tarea recomendada:${RESET}`);
        console.log(`   ${BOLD}${next.id}${RESET} — ${next.title}`);
        console.log(`   Score: ${next.score} | Esfuerzo: ${next.effort}d | Valor: ${next.value}`);
        console.log(`   ${GRAY}${next.notes}${RESET}\n`);
    } else {
        console.log(`${GREEN}🎉 ¡Todas las tareas completadas!${RESET}\n`);
    }
    
    // Estadísticas
    const total = state.taskRegistry.length;
    const done = state.taskRegistry.filter(t => t.status === 'done').length;
    const pending = state.taskRegistry.filter(t => t.status === 'pending').length;
    const inProgress = state.taskRegistry.filter(t => t.status === 'in_progress').length;
    
    console.log(`${GRAY}📊 Progreso: ${done}/${total} completadas | ${pending} pendientes | ${inProgress} en progreso${RESET}`);
    console.log(`${GRAY}📅 Última actualización: ${state.lastUpdated}${RESET}\n`);
}

// Mostrar grafo de dependencias
function showGraph(state) {
    console.log(`\n${BOLD}${CYAN}🔗 GRAFO DE DEPENDENCIAS${RESET}\n`);
    
    for (const task of state.taskRegistry) {
        if (task.status === 'done') continue;
        
        const deps = (task.dependencies || []).join(', ') || '(ninguna)';
        const blocks = state.taskRegistry
            .filter(t => (t.dependencies || []).includes(task.id))
            .map(t => t.id)
            .join(', ') || '(ninguna)';
        
        console.log(`${YELLOW}${task.id}${RESET}`);
        console.log(`   ← deps: ${deps}`);
        console.log(`   → desbloquea: ${blocks}\n`);
    }
}

// Marcar tarea como completada
function markDone(state, taskId) {
    const task = state.taskRegistry.find(t => t.id === taskId);
    if (!task) {
        console.log(`${RED}❌ Tarea no encontrada: ${taskId}${RESET}`);
        return;
    }
    
    // Verificar dependencias
    const depsDone = (task.dependencies || []).every(depId => {
        const dep = state.taskRegistry.find(t => t.id === depId);
        return dep && dep.status === 'done';
    });
    
    if (!depsDone) {
        const pending = (task.dependencies || []).filter(depId => {
            const dep = state.taskRegistry.find(t => t.id === depId);
            return dep && dep.status !== 'done';
        });
        console.log(`${RED}❌ Dependencias pendientes: ${pending.join(', ')}${RESET}`);
        return;
    }
    
    task.status = 'done';
    saveState(state);
    console.log(`${GREEN}✅ Tarea completada: ${task.id} — ${task.title}${RESET}`);
    
    // Mostrar desbloqueos
    const unblocked = state.taskRegistry.filter(t => 
        (t.dependencies || []).includes(taskId) && t.status === 'pending'
    );
    if (unblocked.length > 0) {
        console.log(`${CYAN}🔓 Desbloqueadas:${RESET}`);
        for (const t of unblocked) {
            console.log(`   → ${t.id} — ${t.title}`);
        }
    }
}

// Inicializar con tareas base
function initState(state) {
    if (state.taskRegistry.length > 0) {
        console.log(`${YELLOW}⚠️  El estado ya tiene ${state.taskRegistry.length} tareas.`);
        console.log(`   ¿Sobreescribir? (s/N)${RESET}`);
        // En modo no-interactivo, no sobreescribir
        return;
    }
    saveState(state);
    console.log(`${GREEN}✅ Estado inicializado con ${state.taskRegistry.length} tareas${RESET}`);
}

// Main
const [,, command, ...args] = process.argv;
const state = loadState();

switch (command) {
    case 'status':
        showStatus(state);
        break;
    case 'graph':
        showGraph(state);
        break;
    case 'next':
        showStatus(state);
        break;
    case 'done':
        if (!args[0]) {
            console.log(`${RED}Uso: node scripts/orchestrator-next-steps.js done <TASK_ID>${RESET}`);
            break;
        }
        markDone(state, args[0]);
        break;
    case 'init':
        initState(state);
        break;
    default:
        console.log(`${BOLD}RIDF Orchestrator${RESET}`);
        console.log(`Comandos: status | graph | next | done <id> | init`);
}
