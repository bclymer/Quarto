var Quarto = {};

(function ($) {

	$(document).ready(function () {
		Quarto.main().loadRegisterHTML();
	});

	Quarto.main = (function() {

		function loadGameHTML(uuid) {
			$('body').animate({backgroundColor: "#EEEEEE"}, 500);
			Quarto.game().loadGameHTML(uuid);
		};

		function loadRegisterHTML() {
			Quarto.register().loadRegisterHTML();
		}

		return {
			loadRegisterHTML: loadRegisterHTML,
			loadGameHTML: loadGameHTML
		};
	});

})(jQuery);