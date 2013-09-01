var Quarto = {};

(function ($) {

	$(document).ready(function () {
		Quarto.main().loadRegisterHTML();
	});

	Quarto.main = (function() {

		function loadWaitingRoomHTML() {
			Quarto.waitingRoom().loadWaitingRoom();
		}

		function loadGameHTML() {
			Quarto.game().loadGameHTML();
		};

		function loadRegisterHTML() {
			Quarto.register().loadRegisterHTML();
		}

		return {
			loadWaitingRoomHTML: loadWaitingRoomHTML,
			loadRegisterHTML: loadRegisterHTML,
			loadGameHTML: loadGameHTML,
		};
	});

	$.fn.animateHighlight = function(highlightColor, duration) {
	    var highlightBg = highlightColor || "#FFFF9C";
	    var animateMs = duration || 1500;
	    var originalBg = this.css("backgroundColor");
	    this.stop().css("background-color", highlightBg).animate({backgroundColor: originalBg}, animateMs);
	};

})(jQuery);