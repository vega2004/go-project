// Clase para manejar búsqueda con filtros
class SearchFilters {
    constructor(tableId, options = {}) {
        this.table = document.getElementById(tableId);
        this.options = {
            searchInput: options.searchInput || '#searchInput',
            filterSelects: options.filterSelects || [],
            pagination: options.pagination || true,
            pageSize: options.pageSize || 5,
            ...options
        };
        
        this.currentPage = 1;
        this.totalPages = 1;
        this.data = [];
        this.filteredData = [];
        
        this.init();
    }
    
    init() {
        this.searchInput = document.querySelector(this.options.searchInput);
        this.filterElements = this.options.filterSelects.map(selector => 
            document.querySelector(selector)
        );
        
        this.initEventListeners();
        this.loadData();
    }
    
    initEventListeners() {
        // Búsqueda con debounce
        if (this.searchInput) {
            let timeout;
            this.searchInput.addEventListener('input', () => {
                clearTimeout(timeout);
                timeout = setTimeout(() => {
                    this.currentPage = 1;
                    this.applyFilters();
                }, 300);
            });
        }
        
        // Filtros
        this.filterElements.forEach(filter => {
            if (filter) {
                filter.addEventListener('change', () => {
                    this.currentPage = 1;
                    this.applyFilters();
                });
            }
        });
    }
    
    async loadData() {
        try {
            const response = await API.get(this.options.apiEndpoint || window.location.pathname);
            this.data = response.data || [];
            this.applyFilters();
        } catch (error) {
            console.error('Error cargando datos:', error);
        }
    }
    
    applyFilters() {
        // Aplicar búsqueda
        let filtered = this.data;
        
        if (this.searchInput && this.searchInput.value) {
            const searchTerm = this.searchInput.value.toLowerCase();
            filtered = filtered.filter(item => {
                return Object.values(item).some(val => 
                    String(val).toLowerCase().includes(searchTerm)
                );
            });
        }
        
        // Aplicar filtros
        this.filterElements.forEach((filter, index) => {
            if (filter && filter.value) {
                const filterValue = filter.value.toLowerCase();
                filtered = filtered.filter(item => {
                    const field = this.options.filterFields[index];
                    return String(item[field]).toLowerCase() === filterValue;
                });
            }
        });
        
        this.filteredData = filtered;
        this.totalPages = Math.ceil(filtered.length / this.options.pageSize);
        
        this.renderTable();
        if (this.options.pagination) {
            this.renderPagination();
        }
    }
    
    renderTable() {
        const start = (this.currentPage - 1) * this.options.pageSize;
        const end = start + this.options.pageSize;
        const pageData = this.filteredData.slice(start, end);
        
        const tbody = this.table.querySelector('tbody');
        if (!tbody) return;
        
        if (pageData.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="100%" class="text-center py-4">
                        <i class="bi bi-inbox fs-1 d-block text-muted"></i>
                        No se encontraron resultados
                    </td>
                </tr>
            `;
            return;
        }
        
        // Usar el template de filas existente o generar dinámicamente
        tbody.innerHTML = pageData.map(item => this.renderRow(item)).join('');
    }
    
    renderRow(item) {
        // Este método debe ser sobreescrito según la tabla específica
        console.warn('renderRow debe ser implementado');
        return '';
    }
    
    renderPagination() {
        const paginationEl = document.getElementById('pagination');
        if (!paginationEl) return;
        
        let html = '<ul class="pagination justify-content-center">';
        
        // Botón anterior
        html += `
            <li class="page-item ${this.currentPage === 1 ? 'disabled' : ''}">
                <a class="page-link" href="#" data-page="${this.currentPage - 1}">
                    <i class="bi bi-chevron-left"></i>
                </a>
            </li>
        `;
        
        // Números de página
        for (let i = 1; i <= this.totalPages; i++) {
            if (i === 1 || i === this.totalPages || (i >= this.currentPage - 2 && i <= this.currentPage + 2)) {
                html += `
                    <li class="page-item ${i === this.currentPage ? 'active' : ''}">
                        <a class="page-link" href="#" data-page="${i}">${i}</a>
                    </li>
                `;
            } else if (i === this.currentPage - 3 || i === this.currentPage + 3) {
                html += '<li class="page-item disabled"><span class="page-link">...</span></li>';
            }
        }
        
        // Botón siguiente
        html += `
            <li class="page-item ${this.currentPage === this.totalPages ? 'disabled' : ''}">
                <a class="page-link" href="#" data-page="${this.currentPage + 1}">
                    <i class="bi bi-chevron-right"></i>
                </a>
            </li>
        `;
        
        html += '</ul>';
        
        paginationEl.innerHTML = html;
        
        // Event listeners para paginación
        paginationEl.querySelectorAll('[data-page]').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const page = parseInt(e.target.closest('[data-page]').dataset.page);
                if (page && page !== this.currentPage && page >= 1 && page <= this.totalPages) {
                    this.currentPage = page;
                    this.renderTable();
                    this.renderPagination();
                }
            });
        });
    }
}