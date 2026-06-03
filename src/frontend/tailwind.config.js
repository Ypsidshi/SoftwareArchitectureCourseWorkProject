/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#eff7ff",
          100: "#dbecff",
          200: "#bedcff",
          500: "#2b7fff",
          600: "#1565e0",
          700: "#1150b3",
          900: "#0b2f6b",
        },
      },
    },
  },
  plugins: [],
};
