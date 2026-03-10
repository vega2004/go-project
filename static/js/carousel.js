// Clase para manejar el carrusel interactivo
class ImageCarousel {
    constructor(containerId, options = {}) {
        this.container = document.getElementById(containerId);
        this.options = {
            maxImages: options.maxImages || 10,
            autoPlay: options.autoPlay || true,
            interval: options.interval || 3000,
            ...options
        };
        
        this.images = [];
        this.currentIndex = 0;
        this.autoPlayInterval = null;
        
        this.init();
    }
    
    async init() {
        await this.loadImages();
        this.render();
        this.initEventListeners();
        
        if (this.options.autoPlay) {
            this.startAutoPlay();
        }
    }
    
    async loadImages() {
        try {
            const response = await API.get('/carrusel/api/list');
            this.images = response.imagenes || [];
        } catch (error) {
            console.error('Error cargando imágenes:', error);
            this.showError('No se pudieron cargar las imágenes');
        }
    }
    
    render() {
        if (!this.container) return;
        
        const html = `
            <div class="carousel-container">
                <div class="carousel-main">
                    <button class="carousel-prev" id="prevBtn">
                        <i class="bi bi-chevron-left"></i>
                    </button>
                    
                    <div class="carousel-track-container">
                        <ul class="carousel-track">
                            ${this.images.map((img, index) => `
                                <li class="carousel-slide ${index === 0 ? 'active' : ''}" data-index="${index}">
                                    <img src="${img.ruta}" alt="${img.nombre}" class="carousel-image">
                                    <div class="carousel-caption">
                                        <h4>${img.nombre}</h4>
                                    </div>
                                </li>
                            `).join('')}
                        </ul>
                    </div>
                    
                    <button class="carousel-next" id="nextBtn">
                        <i class="bi bi-chevron-right"></i>
                    </button>
                </div>
                
                <div class="carousel-indicators">
                    ${this.images.map((_, index) => `
                        <span class="indicator ${index === 0 ? 'active' : ''}" data-index="${index}"></span>
                    `).join('')}
                </div>
                
                ${this.isAdmin() ? this.renderUploadForm() : ''}
            </div>
        `;
        
        this.container.innerHTML = html;
        this.updateTrack();
    }
    
    renderUploadForm() {
        return `
            <div class="upload-area mt-4" id="uploadArea">
                <i class="bi bi-cloud-upload"></i>
                <h4>Arrastra y suelta tus imágenes aquí</h4>
                <p>o</p>
                <button class="btn btn-primary" onclick="document.getElementById('fileInput').click()">
                    Seleccionar archivos
                </button>
                <input type="file" id="fileInput" multiple accept="image/*" style="display: none">
                <p class="text-muted small mt-2">Máximo 5MB por imagen. Formatos: JPG, PNG, GIF</p>
            </div>
        `;
    }
    
    initEventListeners() {
        // Botones de navegación
        document.getElementById('prevBtn')?.addEventListener('click', () => this.prev());
        document.getElementById('nextBtn')?.addEventListener('click', () => this.next());
        
        // Indicadores
        document.querySelectorAll('.indicator').forEach(indicator => {
            indicator.addEventListener('click', (e) => {
                const index = parseInt(e.target.dataset.index);
                this.goToSlide(index);
            });
        });
        
        // Upload de imágenes (solo admin)
        const uploadArea = document.getElementById('uploadArea');
        const fileInput = document.getElementById('fileInput');
        
        if (uploadArea && fileInput) {
            this.initUploadHandlers(uploadArea, fileInput);
        }
        
        // Keyboard navigation
        document.addEventListener('keydown', (e) => {
            if (e.key === 'ArrowLeft') this.prev();
            if (e.key === 'ArrowRight') this.next();
        });
    }
    
    initUploadHandlers(uploadArea, fileInput) {
        // Drag & drop
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });
        
        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('dragover');
        });
        
        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            const files = Array.from(e.dataTransfer.files);
            this.handleFiles(files);
        });
        
        // Click para seleccionar
        fileInput.addEventListener('change', (e) => {
            const files = Array.from(e.target.files);
            this.handleFiles(files);
            fileInput.value = ''; // Resetear
        });
    }
    
    async handleFiles(files) {
        // Validar límite de imágenes
        if (this.images.length + files.length > this.options.maxImages) {
            this.showError(`Máximo ${this.options.maxImages} imágenes permitidas`);
            return;
        }
        
        // Validar cada archivo
        for (const file of files) {
            if (!file.type.startsWith('image/')) {
                this.showError(`${file.name} no es una imagen válida`);
                continue;
            }
            
            if (file.size > 5 * 1024 * 1024) {
                this.showError(`${file.name} excede 5MB`);
                continue;
            }
            
            await this.uploadImage(file);
        }
    }
    
    async uploadImage(file) {
        const formData = new FormData();
        formData.append('imagen', file);
        
        try {
            const result = await API.upload('/carrusel/upload', formData);
            await this.loadImages();
            this.render();
            this.showSuccess('Imagen subida exitosamente');
        } catch (error) {
            this.showError(error.message);
        }
    }
    
    async deleteImage(imageId) {
        if (!confirm('¿Estás seguro de eliminar esta imagen?')) return;
        
        try {
            await API.get(`/carrusel/delete/${imageId}`);
            await this.loadImages();
            this.render();
            this.showSuccess('Imagen eliminada');
        } catch (error) {
            this.showError(error.message);
        }
    }
    
    next() {
        if (this.images.length === 0) return;
        this.currentIndex = (this.currentIndex + 1) % this.images.length;
        this.updateTrack();
    }
    
    prev() {
        if (this.images.length === 0) return;
        this.currentIndex = (this.currentIndex - 1 + this.images.length) % this.images.length;
        this.updateTrack();
    }
    
    goToSlide(index) {
        if (index >= 0 && index < this.images.length) {
            this.currentIndex = index;
            this.updateTrack();
        }
    }
    
    updateTrack() {
        const track = document.querySelector('.carousel-track');
        if (!track) return;
        
        const slides = document.querySelectorAll('.carousel-slide');
        const indicators = document.querySelectorAll('.indicator');
        
        slides.forEach((slide, index) => {
            if (index === this.currentIndex) {
                slide.classList.add('active');
            } else {
                slide.classList.remove('active');
            }
        });
        
        indicators.forEach((indicator, index) => {
            if (index === this.currentIndex) {
                indicator.classList.add('active');
            } else {
                indicator.classList.remove('active');
            }
        });
    }
    
    startAutoPlay() {
        this.autoPlayInterval = setInterval(() => {
            this.next();
        }, this.options.interval);
    }
    
    stopAutoPlay() {
        if (this.autoPlayInterval) {
            clearInterval(this.autoPlayInterval);
        }
    }
    
    isAdmin() {
        return document.body.dataset.userRole === 'admin';
    }
    
    showError(message) {
        const alert = this.createAlert('danger', message);
        this.container.prepend(alert);
        setTimeout(() => alert.remove(), 5000);
    }
    
    showSuccess(message) {
        const alert = this.createAlert('success', message);
        this.container.prepend(alert);
        setTimeout(() => alert.remove(), 3000);
    }
    
    createAlert(type, message) {
        const alert = document.createElement('div');
        alert.className = `alert alert-${type} alert-dismissible fade show`;
        alert.innerHTML = `
            <i class="bi ${type === 'success' ? 'bi-check-circle' : 'bi-exclamation-triangle'}"></i>
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        return alert;
    }
}

// Inicializar carrusel cuando el DOM esté listo
document.addEventListener('DOMContentLoaded', () => {
    const carouselElement = document.getElementById('mainCarousel');
    if (carouselElement) {
        window.carousel = new ImageCarousel('mainCarousel', {
            maxImages: 10,
            autoPlay: true,
            interval: 5000
        });
    }
});