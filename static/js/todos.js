// Todo module - handles todo CRUD operations

class TodoManager {
    constructor(auth) {
        this.auth = auth;
        this.todos = [];
        this.filter = 'all';
        this.sortBy = 'priority';
    }

    async fetchTodos(completed = null, limit = 50, offset = 0) {
        let url = `${this.auth.apiBase}/todos?limit=${limit}&offset=${offset}`;
        if (completed !== null) {
            url += `&completed=${completed}`;
        }

        const response = await fetch(url, {
            headers: this.auth.getAuthHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to fetch todos');
        }

        const data = await response.json();
        this.todos = data.data.todos;
        return this.todos;
    }

    async createTodo(todoData) {
        const response = await fetch(`${this.auth.apiBase}/todos`, {
            method: 'POST',
            headers: this.auth.getAuthHeaders(),
            body: JSON.stringify(todoData)
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || 'Failed to create todo');
        }

        this.todos.unshift(data.data);
        return data.data;
    }

    async updateTodo(id, updates) {
        const response = await fetch(`${this.auth.apiBase}/todos?id=${id}`, {
            method: 'PUT',
            headers: this.auth.getAuthHeaders(),
            body: JSON.stringify(updates)
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || 'Failed to update todo');
        }

        const index = this.todos.findIndex(t => t.id === id);
        if (index !== -1) {
            this.todos[index] = { ...this.todos[index], ...updates };
        }

        return true;
    }

    async deleteTodo(id) {
        const response = await fetch(`${this.auth.apiBase}/todos?id=${id}`, {
            method: 'DELETE',
            headers: this.auth.getAuthHeaders()
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || 'Failed to delete todo');
        }

        this.todos = this.todos.filter(t => t.id !== id);
        return true;
    }

    async getStats() {
        const response = await fetch(`${this.auth.apiBase}/todos/stats`, {
            headers: this.auth.getAuthHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to fetch stats');
        }

        const data = await response.json();
        return data.data;
    }

    async getOverdueTodos() {
        const response = await fetch(`${this.auth.apiBase}/todos/overdue`, {
            headers: this.auth.getAuthHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to fetch overdue todos');
        }

        const data = await response.json();
        return data.data;
    }

    async getTodaysTodos() {
        const response = await fetch(`${this.auth.apiBase}/todos/today`, {
            headers: this.auth.getAuthHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to fetch today\'s todos');
        }

        const data = await response.json();
        return data.data;
    }

    getFilteredTodos() {
        let filtered = [...this.todos];

        // Apply status filter
        if (this.filter === 'pending') {
            filtered = filtered.filter(t => !t.completed);
        } else if (this.filter === 'completed') {
            filtered = filtered.filter(t => t.completed);
        }

        // Apply sorting
        filtered.sort((a, b) => {
            switch (this.sortBy) {
                case 'priority':
                    return a.priority - b.priority;
                case 'due-date':
                    if (!a.due_date && !b.due_date) return 0;
                    if (!a.due_date) return 1;
                    if (!b.due_date) return -1;
                    return new Date(a.due_date) - new Date(b.due_date);
                case 'created':
                    return new Date(b.created_at) - new Date(a.created_at);
                default:
                    return 0;
            }
        });

        return filtered;
    }

    formatDate(dateString) {
        if (!dateString) return '';
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    formatDateShort(dateString) {
        if (!dateString) return '';
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric'
        });
    }

    isOverdue(dateString) {
        if (!dateString) return false;
        const dueDate = new Date(dateString);
        const now = new Date();
        return dueDate < now;
    }

    isToday(dateString) {
        if (!dateString) return false;
        const dueDate = new Date(dateString);
        const today = new Date();
        return dueDate.toDateString() === today.toDateString();
    }
}

// Global todo manager
window.todos = null;