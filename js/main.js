(function ($) {

	$(document).ready(function () {
		if (Quarto.username != "") {
			Quarto.socket().makeConnection(Quarto.username, function() {
				Quarto.main().loadWaitingRoomHTML();
			});
		} else {
			Quarto.main().loadRegisterHTML();
		}

		toastr.options = {
            "debug": false,
            "fadeIn": 300,
            "fadeOut": 500,
            "timeOut": 2000,
            "extendedTimeOut": 1000
        }
	});

	$(document).on(Quarto.constants.Info, function (event, data) {
		toastr.info(data.Message);
	});

	$(document).on(Quarto.constants.Error, function (event, data) {
		toastr.error(data.Message);
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