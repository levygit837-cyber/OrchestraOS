/* OrchestraOS Navigation Helper */

const PAGES = {
  dashboard:  { label: 'Dashboard',      icon: 'LayoutGrid',      href: 'index.html' },
  tasks:      { label: 'Tasks',          icon: 'GitBranch',       href: 'task.html?id=task-001' },
  runs:       { label: 'Runs & Logs',    icon: 'Activity',        href: 'run.html?id=run-101' },
  agents:     { label: 'Agents',         icon: 'Bot',             href: 'agents.html' },
  composer:   { label: 'New Task',       icon: 'PenLine',         href: 'composer.html' },
  kanban:     { label: 'Kanban (Beta)',  icon: 'Columns',         href: 'kanban.html' },
};

function renderSidebar(activeKey) {
  const nav = document.getElementById('sidebar-nav');
  if (!nav) return;
  nav.innerHTML = Object.entries(PAGES).map(([key, page]) => {
    const isActive = key === activeKey;
    return `
      <a href="${page.href}"
         class="flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-150
                ${isActive
                  ? 'bg-[var(--accent-dark)] text-[var(--accent-light)] border border-[var(--accent)]'
                  : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]'}">
        <span class="nav-icon" data-icon="${page.icon}"></span>
        <span class="nav-label">${page.label}</span>
      </a>
    `;
  }).join('');
}

function renderHeader(title, breadcrumb) {
  const header = document.getElementById('app-header');
  if (!header) return;
  header.innerHTML = `
    <div class="flex items-center justify-between h-full px-6">
      <div class="flex items-center gap-4">
        <button id="sidebar-toggle" class="p-2 rounded-lg hover:bg-[var(--bg-hover)] text-[var(--text-secondary)] transition-colors">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="18" y2="18"/></svg>
        </button>
        <div class="flex items-center gap-2 text-sm">
          ${breadcrumb.map((b,i) => `
            <span class="${i === breadcrumb.length-1 ? 'text-[var(--text-primary)] font-medium' : 'text-[var(--text-muted)]'}">
              ${i > 0 ? '<span class="mx-2 text-[var(--text-muted)]">/</span>' : ''}${b}
            </span>
          `).join('')}
        </div>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative">
          <input type="text" placeholder="Buscar..." class="w-64 bg-[var(--bg-primary)] border border-[var(--border-color)] rounded-lg px-4 py-1.5 text-sm text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none transition-colors"/>
          <svg class="absolute right-3 top-2 text-[var(--text-muted)]" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
        </div>
        <button class="relative p-2 rounded-lg hover:bg-[var(--bg-hover)] text-[var(--text-secondary)]">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"/><path d="M10.3 21a1.94 1.94 0 0 0 3.4 0"/></svg>
          <span class="absolute top-1 right-1 w-2 h-2 bg-[var(--accent)] rounded-full"></span>
        </button>
      </div>
    </div>
  `;
  document.getElementById('sidebar-toggle').addEventListener('click', () => {
    document.getElementById('app-sidebar').classList.toggle('collapsed');
  });
}
