// Auth module - handles user authentication

class AuthManager {
    constructor() {
        this.token = localStorage.getItem('todoai_token');
        this.user = null;
        this.apiBase = this.getApiBase();
    }

    getApiBase() {
        // Support both development and production
        const host = window.location.hostname;
        if (host === 'localhost' || host === '127.0.0.1') {
            return 'http://localhost:8080/api';
        }
        return window.location.origin + '/api';
    }

    async init() {
        if (this.token) {
            try {
                await this.fetchCurrentUser();
                return true;
            } catch (error) {
                this.logout();
                return false;
            }
        }
        return false;
    }

    async login(username, password) {
        const response = await fetch(`${this.apiBase}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || 'Login failed');
        }

        this.token = data.data.token;
        this.user = data.data.user;
        this.saveToken();
        return this.user;
    }

    async register(username, email, password) {
        const response = await fetch(`${this.apiBase}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, email, password })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || 'Registration failed');
        }

        this.token = data.data.token;
        this.user = data.data.user;
        this.saveToken();
        return this.user;
    }

    async fetchCurrentUser() {
        if (!this.token) throw new Error('No token');

        const response = await fetch(`${this.apiBase}/auth/me`, {
            headers: {
                'Authorization': `Bearer ${this.token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to fetch user');
        }

        const data = await response.json();
        this.user = data.data;
        return this.user;
    }

    logout() {
        this.token = null;
        this.user = null;
        localStorage.removeItem('todoai_token');
    }

    saveToken() {
        localStorage.setItem('todoai_token', this.token);
    }

    isAuthenticated() {
        return !!this.token && !!this.user;
    }

    getAuthHeaders() {
        return {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
        };
    }

    // Refresh token if needed (for future implementation)
    async refreshToken() {
        // TODO: Implement token refresh
        console.log('Token refresh not implemented');
    }
}

// Global auth instance
window.auth = new AuthManager();