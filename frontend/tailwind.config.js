/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Логотип
        'logo-bright': '#e50102',
        'logo-dark': '#9e0202',
        
        // Тёмная тема
        'bg-dark': '#05022a',
        'bg-gradient': '#0f2667',
        'card-dark': '#1b2a50',
        'text-dark': '#d4d8e8',
        
        // Светлая тема
        'bg-light': '#faf4f2',
        'card-light': '#ffffff',
        'text-light': '#333333',
        'link-light': '#c62828',
        'link-dark': '#8e1c1c',
        'accent-light': '#ef6c00',
        'accent-dark': '#c24f00',
      },
    },
  },
  plugins: [],
}
