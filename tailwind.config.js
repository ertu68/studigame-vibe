/** @type {import('tailwindcss').Config} */
export default {
	content: [
		"./internal/web/templates/**/*.html",
		"./internal/web/components/**/*.html",
	],
	theme: {
		extend: {
			colors: {
				asphalt: "#1C1B1A",
				pfandglas: "#7A9B57",
				verkehrsorange: "#D9711F",
				muenzgold: "#C9A24B",
				pappe: "#E8E1D3",
				backstein: "#A83E32",
			},
			fontFamily: {
				display: ["ui-sans-serif", "system-ui", "Segoe UI", "Roboto", "sans-serif"],
				body: ["ui-sans-serif", "system-ui", "Segoe UI", "Roboto", "sans-serif"],
			},
		},
	},
	plugins: [],
};
