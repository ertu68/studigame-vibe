// Live countdown for the running action on /actions.
// Loaded as a 'self'-served file so the Content-Security-Policy (script-src
// 'self') allows it. The alpine:init listener runs before Alpine initialises;
// the meta refresh stays as the no-JS fallback. Reloads when the shift is due
// so the server settles it.
document.addEventListener("alpine:init", () => {
	Alpine.data("countdown", (endsAtIso) => ({
		end: new Date(endsAtIso),
		now: new Date(),
		init() {
			const tick = () => {
				if (this.left <= 0) {
					location.reload();
					return;
				}
				this.now = new Date();
			};
			tick();
			setInterval(tick, 1000);
		},
		get left() {
			return Math.max(0, this.end - this.now);
		},
		get expired() {
			return this.left <= 0;
		},
		get display() {
			const t = Math.floor(this.left / 1000);
			const h = Math.floor(t / 3600);
			const m = Math.floor((t % 3600) / 60);
			const s = t % 60;
			const mm = String(m).padStart(2, "0");
			const ss = String(s).padStart(2, "0");
			return h > 0 ? h + ":" + mm + ":" + ss : mm + ":" + ss;
		},
	}));
});
