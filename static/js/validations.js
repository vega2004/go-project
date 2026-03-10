// Validaciones en frontend
const Validator = {
    // Validar nombre (solo letras y espacios)
    name: (value) => {
        const regex = /^[a-zA-ZáéíóúÁÉÍÓÚñÑüÜ\s'-]{2,50}$/;
        return regex.test(value);
    },
    
    // Validar email
    email: (value) => {
        const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        return regex.test(value);
    },
    
    // Validar teléfono
    phone: (value) => {
        // Remover espacios y contar dígitos
        const digits = value.replace(/\D/g, '');
        return digits.length >= 8 && digits.length <= 15;
    },
    
    // Validar contraseña
    password: (value) => {
        return value.length >= 6;
    },
    
    // Validar que dos campos coincidan
    match: (value1, value2) => {
        return value1 === value2;
    },
    
    // Validar formulario completo
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
            
            // Validar match si existe
            if (input.dataset.match) {
                const matchInput = document.querySelector(input.dataset.match);
                if (matchInput && !Validator.match(value, matchInput.value)) {
                    errors.push(input.dataset.matchError || 'Los campos no coinciden');
                }
            }
        });
        
        return {
            isValid: errors.length === 0,
            errors: errors
        };
    },
    
    // Mostrar errores en el formulario
    showErrors: (formElement, errors) => {
        // Limpiar errores anteriores
        formElement.querySelectorAll('.invalid-feedback').forEach(el => el.remove());
        formElement.querySelectorAll('.is-invalid').forEach(el => {
            el.classList.remove('is-invalid');
        });
        
        // Mostrar nuevos errores
        const errorDiv = document.createElement('div');
        errorDiv.className = 'alert alert-danger';
        errorDiv.innerHTML = `
            <i class="bi bi-exclamation-triangle"></i>
            <ul class="mb-0">
                ${errors.map(e => `<li>${e}</li>`).join('')}
            </ul>
        `;
        
        formElement.prepend(errorDiv);
        
        // Auto-remover después de 5 segundos
        setTimeout(() => errorDiv.remove(), 5000);
    }
};

// Inicializar validaciones en formularios
document.addEventListener('DOMContentLoaded', () => {
    const forms = document.querySelectorAll('[data-validate-form]');
    
    forms.forEach(form => {
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            
            const result = Validator.validateForm(form);
            
            if (result.isValid) {
                // Si hay reCAPTCHA, procesarlo
                const recaptcha = form.querySelector('.g-recaptcha');
                if (recaptcha) {
                    const token = grecaptcha.getResponse();
                    if (!token) {
                        alert('Por favor, complete la verificación reCAPTCHA');
                        return;
                    }
                    document.getElementById('recaptchaToken').value = token;
                }
                
                form.submit();
            } else {
                Validator.showErrors(form, result.errors);
            }
        });
        
        // Validación en tiempo real
        const inputs = form.querySelectorAll('[data-validate]');
        inputs.forEach(input => {
            input.addEventListener('blur', () => {
                const result = Validator.validateForm(form);
                if (!result.isValid) {
                    input.classList.add('is-invalid');
                } else {
                    input.classList.remove('is-invalid');
                }
            });
        });
    });
});