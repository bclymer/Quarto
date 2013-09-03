var Quarto = {};

(function ($) {

	$(document).ready(function () {
		Quarto.main().loadRegisterHTML();
	});

	Quarto.main = (function() {

		unloadLoadedModules();

		function loadWaitingRoomHTML() {
			Quarto.waitingRoom().start();
		}

		function loadGameHTML() {
			Quarto.game().start();
		};

		function loadRegisterHTML() {
			Quarto.register().start();
		}

		return {
			loadWaitingRoomHTML: loadWaitingRoomHTML,
			loadRegisterHTML: loadRegisterHTML,
			loadGameHTML: loadGameHTML,
		};
	});

	function unloadLoadedModules() {
		if (Quarto.chat().isLoaded()) {
			Quarto.chat().stop();
		}
		if (Quarto.register().isLoaded()) {
			Quarto.register().stop();
		}
		if (Quarto.waitingRoom().isLoaded()) {
			Quarto.waitingRoom().stop();
		}
		if (Quarto.game().isLoaded()) {
			Quarto.game().stop();
		}
		if (Quarto.gameUi().isLoaded()) {
			Quarto.gameUi().stop();
		}
	}

})(jQuery);