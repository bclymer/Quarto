(function ($) {

	var loaded = false;

	Quarto.register = (function () {
		function start() {
			loaded = true;
			console.log("register start()");
			$('#register').show();
			var username = $("#username");

			username.keydown(function (e) {
				if (e.which == 13 && username.val().length > 0) {
					$('#submit-username').click();
					return false;
				}
			});
			username.keyup(function (e) {
				if (username.val().length > 0) {
					$('#submit-username').removeClass('disabled');
				} else {
					$('#submit-username').addClass('disabled');
				}
			});
			username.focus();
			if (username.val().length > 0) {
				$('#submit-username').removeClass('disabled');
			} else {
				$('#submit-username').addClass('disabled');
			}
			$('#register').hide();
			$('#register').slideDown(500);
			$('#submit-username').on("click", function() {
				$.get("validate?username=" + username.val(),
					function(data) {
						if (data.Valid) {
							Quarto.socket().makeConnection(username.val(), function() {
								Quarto.main().loadWaitingRoomHTML();
							});
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