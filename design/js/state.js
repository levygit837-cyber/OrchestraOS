/* OrchestraOS Mock Data / Shared State */

const AppState = {
  tasks: [
    { id: 'task-001', title: 'Implementar Event Store', status: 'running', priority: 'high', agentCount: 2, createdAt: '2026-05-15T10:00:00Z' },
    { id: 'task-002', title: 'Refatorar Task Graph Planner', status: 'pending', priority: 'medium', agentCount: 1, createdAt: '2026-05-15T11:30:00Z' },
    { id: 'task-003', title: 'Integrar Sandbox Manager', status: 'completed', priority: 'high', agentCount: 3, createdAt: '2026-05-14T09:15:00Z' },
    { id: 'task-004', title: 'Policy Engine v1', status: 'paused', priority: 'low', agentCount: 1, createdAt: '2026-05-13T14:00:00Z' },
  ],
  agents: [
    { id: 'agent-01', name: 'Codex-Builder', profile: 'builder', capabilities: ['code', 'refactor', 'test'], status: 'running', currentTask: 'task-001', sessionId: 'sess-101' },
    { id: 'agent-02', name: 'Orchestrator-Main', profile: 'orchestrator', capabilities: ['plan', 'decompose', 'monitor'], status: 'running', currentTask: 'task-001', sessionId: 'sess-102' },
    { id: 'agent-03', name: 'Review-Bot', profile: 'reviewer', capabilities: ['review', 'diff', 'validate'], status: 'idle', currentTask: null, sessionId: null },
    { id: 'agent-04', name: 'Doc-Writer', profile: 'writer', capabilities: ['document', 'adr', 'test-spec'], status: 'idle', currentTask: null, sessionId: null },
  ],
  workUnits: {
    'task-001': [
      { id: 'wu-001', title: 'Criar schema de eventos', objective: 'Definir tabelas e índices', status: 'completed', agentId: 'agent-01', dependsOn: [] },
      { id: 'wu-002', title: 'Implementar envelope', objective: 'Normalizar payload e metadados', status: 'running', agentId: 'agent-01', dependsOn: ['wu-001'] },
      { id: 'wu-003', title: 'Adicionar replay', objective: 'Reconstruir estado a partir de eventos', status: 'pending', agentId: 'agent-01', dependsOn: ['wu-002'] },
      { id: 'wu-004', title: 'Validar E2E', objective: 'Cenário completo task → event → query', status: 'pending', agentId: 'agent-02', dependsOn: ['wu-003'] },
    ],
    'task-003': [
      { id: 'wu-010', title: 'Criar branch por task', objective: 'Isolamento via git worktree', status: 'completed', agentId: 'agent-03', dependsOn: [] },
      { id: 'wu-011', title: 'Container Docker', objective: 'Imagem base com Go e dependências', status: 'completed', agentId: 'agent-04', dependsOn: [] },
      { id: 'wu-012', title: 'Mount de volume', objective: 'Mapear worktree no container', status: 'completed', agentId: 'agent-03', dependsOn: ['wu-010','wu-011'] },
    ]
  },
  runs: [
    { id: 'run-101', taskId: 'task-001', workUnitId: 'wu-002', status: 'running', agentSessionId: 'sess-101', startedAt: '2026-05-15T10:30:00Z', events: 47 },
    { id: 'run-102', taskId: 'task-001', workUnitId: 'wu-004', status: 'pending', agentSessionId: null, startedAt: null, events: 0 },
    { id: 'run-201', taskId: 'task-003', workUnitId: 'wu-012', status: 'completed', agentSessionId: 'sess-201', startedAt: '2026-05-14T09:30:00Z', events: 23 },
  ],
  logs: [
    { ts: '2026-05-15T10:30:01Z', level: 'info', source: 'orchestrator', message: 'Run run-101 iniciada para wu-002' },
    { ts: '2026-05-15T10:30:05Z', level: 'info', source: 'agent-01', message: 'Checkpoint #1: schema validado' },
    { ts: '2026-05-15T10:31:12Z', level: 'info', source: 'agent-01', message: 'Tool request: file_write (events.go)' },
    { ts: '2026-05-15T10:31:15Z', level: 'warn', source: 'policy', message: 'Tool request aguardando aprovação humana' },
    { ts: '2026-05-15T10:32:00Z', level: 'success', source: 'policy', message: 'Tool request aprovada por operator' },
    { ts: '2026-05-15T10:32:30Z', level: 'info', source: 'agent-01', message: 'Arquivo events.go escrito (142 linhas)' },
    { ts: '2026-05-15T10:33:00Z', level: 'info', source: 'agent-01', message: 'Checkpoint #2: implementação concluída' },
    { ts: '2026-05-15T10:33:45Z', level: 'info', source: 'orchestrator', message: 'Propagando evento para subscribers' },
    { ts: '2026-05-15T10:34:00Z', level: 'error', source: 'runtime', message: 'Heartbeat sess-101 atrasado (3.2s)' },
    { ts: '2026-05-15T10:34:05Z', level: 'info', source: 'runtime', message: 'Heartbeat recuperado' },
  ]
};

/* Helpers */
function getStatusColor(status) {
  const map = {
    running: 'var(--status-running)',
    completed: 'var(--status-completed)',
    failed: 'var(--status-failed)',
    paused: 'var(--status-paused)',
    pending: 'var(--status-pending)',
    idle: 'var(--status-pending)'
  };
  return map[status] || map.pending;
}

function getAgentById(id) {
  return AppState.agents.find(a => a.id === id);
}

function getTaskById(id) {
  return AppState.tasks.find(t => t.id === id);
}

function getWorkUnitsForTask(taskId) {
  return AppState.workUnits[taskId] || [];
}
