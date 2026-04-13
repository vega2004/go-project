// API Service - Manejo de peticiones con Fetch API
const API = {
    baseURL: '',
    
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    },

    async get(endpoint) {
        try {
            const response = await fetch(this.baseURL + endpoint, {
                method: 'GET',
                headers: this.headers,
                credentials: 'same-origin'
            });
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            return await response.json();
        } catch (error) {
            console.error('API GET Error:', error);
            throw error;
        }
    },

    async post(endpoint, data) {
        try {
            const response = await fetch(this.baseURL + endpoint, {
                method: 'POST',
                headers: this.headers,
                credentials: 'same-origin',
                body: JSON.stringify(data)
            });
            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Error en la petición');
            }
            return await response.json();
        } catch (error) {
            console.error('API POST Error:', error);
            throw error;
        }
    },

    async delete(endpoint) {
        try {
            const response = await fetch(this.baseURL + endpoint, {
                method: 'DELETE',
                headers: this.headers,
                credentials: 'same-origin'
            });
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            return await response.json();
        } catch (error) {
            console.error('API DELETE Error:', error);
            throw error;
        }
    },

    async upload(endpoint, formData) {
        try {
            const response = await fetch(this.baseURL + endpoint, {
                method: 'POST',
                credentials: 'same-origin',
                body: formData
            });
            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Error al subir archivo');
            }
            return await response.json();
        } catch (error) {
            console.error('API Upload Error:', error);
            throw error;
        }
    }
};