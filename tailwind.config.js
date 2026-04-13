/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/templates/**/*.html",
    "./web/templates/auth/**/*.html",
    "./web/static/js/**/*.js",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#E6F0FA',
          100: '#CCE0F5',
          200: '#99C2EB',
          300: '#66A3E0',
          400: '#3385D6',
          500: '#0A66CC',
          600: '#0852A3',
          700: '#063D7A',
          800: '#042952',
          900: '#021429',
        },
        secondary: {
          500: '#6C757D',
          700: '#495057',
        },
        success: '#10B981',
        warning: '#F59E0B',
        error: '#EF4444',
        info: '#3B82F6',
        // Colores personalizados para el fondo
        dark: '#021024',
        'primary-dark': '#052659',
      },
    },
  },
  plugins: [],
}