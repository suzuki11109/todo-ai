// Main app module - orchestrates UI and business logic

class TodoApp {
    constructor() {
        this.auth = window.auth;
        this.todoManager = null;
        this.stats = null;
        this.showDetails = false;
    }

    async init() {
        this.bindEvents();
        const isAuthenticated = await this.auth.init();

        if (isAuthenticated) {
            this.showMainApp();
        } else {
            this.showAuth();
        }
    }

    bindEvents() {
        // Auth form switching
        document.getElementById('show-register').addEventListener('click', (e) => {
            e.preventDefault();
            this.showForm('register');
        });

        document.getElementById('show-login').addEventListener('click', (e) => {
            e.preventDefault();
            this.showForm('login');
        });

        // Login form
        document.getElementById('login-form-element').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleLogin();
        });

        // Register form
        document.getElementById('register-form-element').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleRegister();
        });

        // Logout
        document.getElementById('logout-btn').addEventListener('click', () => this.handleLogout());

        // Add todo - quick add
        document.getElementById('add-todo-form').addEventListener('submit', async (e) => {
            if (this.showDetails) {
                e.preventDefault();
                await this.handleAddTodoWithDetails();
            } else {
                e.preventDefault();
                this.showTodoDetails();
            }
        });

        // Cancel todo details
        document.getElementById('cancel-todo-btn').addEventListener('click', () => {
            this.hideTodoDetails();
        });

        // Quick filters
        document.getElementById('show-today-btn').addEventListener('click', () => this.setFilter('today'));
        document.getElementById('show-overdue-btn').addEventListener('click', () => this.setFilter('overdue'));
        document.getElementById('show-all-btn').addEventListener('click', () => this.setFilter('all'));

        // Filter and sort
        document.getElementById('status-filter').addEventListener('change', (e) => {
            this.todoManager.filter = e.target.value;
            this.renderTodos();
        });

        document.getElementById('sort-select').addEventListener('change', (e) => {
            this.todoManager.sortBy = e.target.value;
            this.renderTodos();
        });
    }

    showForm(form) {
        document.getElementById('login-form').classList.remove('active');
        document.getElementById('register-form').classList.remove('active');
        document.getElementById(`${form}-form`).classList.add('active');
    }

    showAuth() {
        document.getElementById('auth-container').style.display = 'block';
        document.getElementById('main-app').classList.add('hidden');
    }

    showMainApp() {
        document.getElementById('auth-container').style.display = 'none';
        document.getElementById('main-app').classList.remove('hidden');

        this.todoManager = new TodoManager(this.auth);
        this.loadData();
    }

    async handleLogin() {
        const form = document.getElementById('login-form-element');
        const submitBtn = form.querySelector('button[type="submit"]');
        const username = form.username.value;
        const password = form.password.value;

        this.setLoading(submitBtn, true);
        this.clearErrors();

        try {
            await this.auth.login(username, password);
            this.showMainApp();
            this.showToast('Welcome back!', 'success');
        } catch (error) {
            this.showError('login-error', error.message);
        } finally {
            this.setLoading(submitBtn, false);
        }
    }

    async handleRegister() {
        const form = document.getElementById('register-form-element');
        const submitBtn = form.querySelector('button[type="submit"]');
        const username = form.username.value;
        const email = form.email.value;
        const password = form.password.value;

        this.setLoading(submitBtn, true);
        this.clearErrors();

        try {
            await this.auth.register(username, email, password);
            this.showMainApp();
            this.showToast('Account created successfully!', 'success');
        } catch (error) {
            this.showError('register-error', error.message);
        } finally {
            this.setLoading(submitBtn, false);
        }
    }

    handleLogout() {
        this.auth.logout();
        this.showAuth();
        this.showToast('Logged out successfully', 'success');
    }

    showTodoDetails() {
        this.showDetails = true;
        document.getElementById('todo-details').style.display = 'block';
    }

    hideTodoDetails() {
        this.showDetails = false;
        document.getElementById('todo-title').value = '';
        document.getElementById('todo-due-date').value = '';
        document.getElementById('todo-notes').value = '';
        document.getElementById('todo-details').style.display = 'none';
    }

    async handleAddTodoWithDetails() {
        const title = document.getElementById('todo-title').value.trim();
        const priority = parseInt(document.getElementById('todo-priority').value);
        const dueDate = document.getElementById('todo-due-date').value || null;
        const notes = document.getElementById('todo-notes').value.trim();

        if (!title) {
            this.showToast('Please enter a task title', 'error');
            return;
        }

        try {
            await this.todoManager.createTodo({
                title,
                priority,
                due_date: dueDate,
                notes
            });

            this.hideTodoDetails();
            this.showToast('Task added!', 'success');
            await this.loadData();
        } catch (error) {
            this.showToast('Failed to add task: ' + error.message, 'error');
        }
    }

    setFilter(filter) {
        // Update active state on chips
        document.querySelectorAll('.action-chip').forEach(chip => chip.classList.remove('active'));

        switch (filter) {
            case 'today':
                this.todoManager.filter = 'pending';
                document.getElementById('show-today-btn').classList.add('active');
                break;
            case 'overdue':
                this.todoManager.filter = 'pending';
                document.getElementById('show-overdue-btn').classList.add('active');
                break;
            default:
                this.todoManager.filter = 'all';
                document.getElementById('show-all-btn').classList.add('active');
                break;
        }

        this.currentView = filter;
        this.renderTodos();
    }

    async loadData() {
        try {
            // Load todos and stats in parallel
            const [todos, stats] = await Promise.all([
                this.todoManager.fetchTodos(),
                this.todoManager.getStats()
            ]);

            this.stats = stats;
            this.renderStats();
            this.renderTodos();
        } catch (error) {
            console.error('Failed to load data:', error);
            this.showToast('Failed to load data', 'error');
        }
    }

    renderStats() {
        if (!this.stats) return;

        document.getElementById('stat-total').textContent = this.stats.total;
        document.getElementById('stat-pending').textContent = this.stats.pending;
        document.getElementById('stat-completed').textContent = this.stats.completed;
        document.getElementById('stat-overdue').textContent = this.stats.overdue;

        const progress = document.getElementById('completion-progress');
        progress.style.width = `${this.stats.completion_rate}%`;

        document.getElementById('completion-text').textContent =
            `${this.stats.completion_rate.toFixed(1)}% completed`;
    }

    renderTodos() {
        let todos = this.todoManager.getFilteredTodos();

        // Apply view filtering for today/overdue
        if (this.currentView === 'today') {
            todos = todos.filter(t => this.todoManager.isToday(t.due_date) && !t.completed);
        } else if (this.currentView === 'overdue') {
            todos = todos.filter(t => this.todoManager.isOverdue(t.due_date) && !t.completed);
        }

        const container = document.getElementById('todos-container');
        const emptyState = document.getElementById('empty-state');

        if (todos.length === 0) {
            container.innerHTML = '';
            emptyState.classList.add('show');
            return;
        }

        emptyState.classList.remove('show');
        container.innerHTML = todos.map(todo => this.renderTodoItem(todo)).join('');

        // Attach event listeners
        container.querySelectorAll('.todo-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', async (e) => {
                const id = e.target.dataset.id;
                await this.toggleTodoComplete(id, e.target.checked);
            });
        });

        container.querySelectorAll('.edit-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const id = e.currentTarget.dataset.id;
                await this.editTodo(id);
            });
        });

        container.querySelectorAll('.delete-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const id = e.currentTarget.dataset.id;
                await this.deleteTodo(id);
            });
        });
    }

    renderTodoItem(todo) {
        const overdue = todo.due_date && this.todoManager.isOverdue(todo.due_date) && !todo.completed;
        const priorityClass = ['', 'high', 'medium', 'low'][todo.priority] || 'medium';
        const priorityLabel = ['', 'High', 'Medium', 'Low'][todo.priority] || 'Medium';

        return `
            <div class="todo-item ${todo.completed ? 'completed' : ''} ${overdue ? 'overdue' : ''}">
                <input type="checkbox"
                       class="todo-checkbox"
                       data-id="${todo.id}"
                       ${todo.completed ? 'checked' : ''}>
                <div class="todo-content">
                    <div class="todo-title">${this.escapeHtml(todo.title)}</div>
                    ${todo.notes ? `<div class="todo-notes">${this.escapeHtml(todo.notes)}</div>` : ''}
                    <div class="todo-meta">
                        <span class="priority-badge ${priorityClass}">${priorityLabel}</span>
                        ${todo.due_date ?
                            `<span class="due-date ${overdue ? 'overdue' : ''}">
                                📅 ${this.todoManager.formatDateShort(todo.due_date)}
                            </span>` : ''
                        }
                        <span>Created: ${this.todoManager.formatDateShort(todo.created_at)}</span>
                    </div>
                </div>
                <div class="todo-actions">
                    <button class="action-btn edit-btn" data-id="${todo.id}" title="Edit">✏️</button>
                    <button class="action-btn delete delete-btn" data-id="${todo.id}" title="Delete">🗑️</button>
                </div>
            </div>
        `;
    }

    async toggleTodoComplete(id, completed) {
        try {
            await this.todoManager.updateTodo(id, { completed });
            await this.loadData();
        } catch (error) {
            this.showToast('Failed to update task', 'error');
        }
    }

    async editTodo(id) {
        const todo = this.todoManager.todos.find(t => t.id === id);
        if (!todo) return;

        // For now, use prompt for simple editing (could be enhanced with modal)
        const newTitle = prompt('Edit task:', todo.title);
        if (!newTitle || newTitle === todo.title) return;

        try {
            await this.todoManager.updateTodo(id, { title: newTitle });
            this.showToast('Task updated', 'success');
            await this.loadData();
        } catch (error) {
            this.showToast('Failed to update task', 'error');
        }
    }

    async deleteTodo(id) {
        if (!confirm('Are you sure you want to delete this task?')) return;

        try {
            await this.todoManager.deleteTodo(id);
            this.showToast('Task deleted', 'success');
            await this.loadData();
        } catch (error) {
            this.showToast('Failed to delete task', 'error');
        }
    }

    setLoading(button, loading) {
        if (loading) {
            button.classList.add('loading');
            button.disabled = true;
        } else {
            button.classList.remove('loading');
            button.disabled = false;
        }
    }

    showError(elementId, message) {
        const errorEl = document.getElementById(elementId);
        if (errorEl) {
            errorEl.textContent = message;
            errorEl.style.display = 'block';
            setTimeout(() => {
                errorEl.style.display = 'none';
            }, 3000);
        }
    }

    clearErrors() {
        document.querySelectorAll('.error-message').forEach(el => {
            el.style.display = 'none';
        });
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toast-container');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;

        container.appendChild(toast);

        setTimeout(() => {
            toast.style.opacity = '0';
            toast.style.transform = 'translateX(100%)';
            toast.style.transition = 'all 0.3s ease';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const app = new TodoApp();
    app.init();
});