(function ($) {

	var loaded = false;
	var roomName;
	var playerOne;
	var playerTwo;
	var private;
	var observers;
	
	Quarto.gameUi = (function () {

		function start() {
			loaded = true;
			console.log("gameUi start()");
			roomName = $('#room-name');
			playerOne = $('#player-one');
			playerTwo = $('#player-two');
			private = $('#private');
			observers = $('#observers');

            $(document).on(Quarto.constants.changeRoom, function (event, data) {
                roomName.text(data.Name);
                if (!data.PlayerOne) {
                	playerOne.html('<button type="button" class="btn btn-primary" id="request-player-one">Dibs!</button>');
                } else if (data.PlayerOne == Quarto.socket().getUsername()) {
                	playerOne.html('<button type="button" class="btn btn-danger" id="leave-player-one">You, Leave?</button>')
                } else {
                	playerOne.text(data.PlayerOne);
                }
                if (!data.PlayerTwo) {
                	playerTwo.html('<button type="button" class="btn btn-primary" id="request-player-two">Dibs!</button>');
                } else if (data.PlayerTwo == Quarto.socket().getUsername()) {
                	playerTwo.html('<button type="button" class="btn btn-danger" id="leave-player-two">You, Leave?</button>')
                } else {
                	playerTwo.text(data.PlayerTwo);
                }
                $('#private').text(data.Private ? "Yes" : "No");
                observers.empty();
                $(data.Observers).each(function(index, observer) {
                	observers.append('<li>' + observer + '</li>')
                });
            });
		}

		function stop() {
			roomName = undefined;
			playerOne = undefined;
			playerTwo = undefined;
			private = undefined;
			observers = undefined;
			$(document).off(Quarto.constants.changeRoom);
			loaded = false;
			console.log("gameUi stop()");
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