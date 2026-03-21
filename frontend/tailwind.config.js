/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#a32d2d',
        'primary-hover': '#8b2020',
        secondary: '#faece7',
        bg: '#F5F4F0',
        surface: '#ffffff',
        text: '#18180f',
        'text-muted': '#6b6a62',
        safe: '#639922',
        match: '#378add',
        reach: '#ba7517'
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
