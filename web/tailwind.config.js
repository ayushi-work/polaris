/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        serif: ['"Instrument Serif"', "Georgia", "serif"],
        sans: [
          "Inter",
          "-apple-system",
          "BlinkMacSystemFont",
          '"Segoe UI"',
          "Roboto",
          "sans-serif",
        ],
        mono: [
          '"JetBrains Mono"',
          "Menlo",
          "Monaco",
          "monospace",
        ],
      },
      colors: {
        border: "hsl(30 10% 88%)",
        ring: "hsl(20 8% 65%)",
        background: "hsl(30 20% 98%)",
        foreground: "hsl(30 10% 15%)",
        primary: {
          DEFAULT: "hsl(220 15% 20%)",
          foreground: "hsl(30 20% 98%)",
        },
        secondary: {
          DEFAULT: "hsl(30 10% 94%)",
          foreground: "hsl(30 10% 25%)",
        },
        destructive: {
          DEFAULT: "hsl(0 60% 45%)",
          foreground: "hsl(0 0% 98%)",
        },
        muted: {
          DEFAULT: "hsl(30 10% 94%)",
          foreground: "hsl(30 5% 50%)",
        },
        accent: {
          DEFAULT: "hsl(30 10% 94%)",
          foreground: "hsl(30 10% 20%)",
        },
        card: {
          DEFAULT: "hsl(0 0% 100%)",
          foreground: "hsl(30 10% 15%)",
        },
      },
      borderRadius: {
        lg: "0.625rem",
        md: "0.5rem",
        sm: "0.375rem",
      },
    },
  },
  plugins: [],
};
