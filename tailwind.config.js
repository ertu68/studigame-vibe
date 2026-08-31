/** @type {import('tailwindcss').Config} */
// Tailwind v4: content paths live here, colors/fonts are defined as @theme
// tokens in static/css/input.css (backed by runtime CSS variables so the
// light/dark theme can flip them).
export default {
	content: [
		"./internal/web/templates/**/*.html",
		"./internal/web/components/**/*.html",
	],
};
