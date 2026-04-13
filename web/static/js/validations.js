// Validaciones en frontend
const Validator = {
    name: (value) => {
        const regex = /^[a-zA-ZáéíóúÁÉÍÓÚñÑüÜ\s'-]{2,50}$/;
        return regex.test(value);
    },
    
    email: (value) => {
        const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        return regex.test(value);
    },
    
    phone: (value) => {
        const digits = value.replace(/\D/g, '');
        return digits.length >= 8 && digits.length <= 15;
    },
    
    password: (value) => value.length >= 6,
    
    match: (value1, value2) => value1 === value2,
    
    validateForm: (formElement) => {
        const errors = [];
        const inputs = formElement.querySelectorAll('[data-validate]');
        
        inputs.forEach(input => {
            const rules = input.dataset.validate.split(' ');
            const value = input.value.trim();
            
            rules.forEach(rule => {
                if (rule === 'required' && !value) {
                    errors.push(`${input.dataset.label || 'Este campo'} es requerido`);
                } else if (rule === 'name' && !Validator.name(value)) {
                    errors.push(`${input.dataset.label || 'El nombre'} no es válido`);
                } else if (rule === 'email' && !Validator.email(value)) {
                    errors.push('Email no válido');
                } else if (rule === 'phone' && !Validator.phone(value)) {
                    errors.push('Teléfono no válido');
                } else if (rule === 'password' && !Validator.password(value)) {
                    errors.push('La contraseña debe tener al menos 6 caracteres');
                }
            });
            
            if (input.dataset.match) {
                const matchInput = document.querySelector(input.dataset.match);
                if (matchInput && !Validator.match(value, matchInput.value)) {
                    errors.push(input.dataset.matchError || 'Los campos no coinciden');
                }
            }
        });
        
        return { isValid: errors.length === 0, errors };
    },
    
    showErrors: (formElement, errors) => {
        formElement.querySelectorAll('.invalid-feedback').forEach(el => el.remove());
        formElement.querySelectorAll('.is-invalid').forEach(el => el.classList.remove('is-invalid'));
        
        const errorDiv = document.createElement('div');
        errorDiv.className = 'alert alert-danger';
        errorDiv.innerHTML = `<i class="bi bi-exclamation-triangle"></i><ul class="mb-0">${errors.map(e => `<li>${e}</li>`).join('')}</ul>`;
        formElement.prepend(errorDiv);
        setTimeout(() => errorDiv.remove(), 5000);
    }
};

// Inicializar validaciones
document.addEventListener('DOMContentLoaded', () => {
    const forms = document.querySelectorAll('[data-validate-form]');
    forms.forEach(form => {
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            const result = Validator.validateForm(form);
            if (result.isValid) {
                form.submit();
            } else {
                Validator.showErrors(form, result.errors);
            }
        });
    });
});