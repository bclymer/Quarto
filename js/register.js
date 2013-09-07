(function ($) {

	var loaded = false;

	Quarto.register = (function () {
		function start() {
			loaded = true;
			console.log("register start()");
			$('#register').show();
			var username = $("#username");

			username.keydown(function (e) {
				if (e.which == 13) {
					$('#submit-username').click();
					return false;
				}
			});
			username.focus();

			$('#register').hide();
			$('#register').slideDown(500);
			$('#submit-username').on("click", function() {
				$.get("validate?uuid=" + username.val(),
					function(data) {
						if (data.Valid) {
							Quarto.socket().makeConnection(username.val());
							Quarto.main().loadWaitingRoomHTML();
						} else {
							toastr.error('Username "' + username.val() + '" is already in use.', 'Error')
						}
					},
					"json"
				);
			});
		}

		function stop() {
			$('#submit-username').off('click')
			$('#register').hide();
			loaded = false;
			console.log("register stop()");
		}

		function isLoaded() {
			return loaded;
		}

		return {
			start: start,
			stop: stop,
			isLoaded: isLoaded,
		}
	});

})(jQuery);