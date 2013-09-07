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

            $(document).on(Quarto.constants.RoomChange, function (event, data) {
                roomName.text(data.Name);
                if (!data.PlayerOne) {
                	playerOne.html('<button type="button" class="btn btn-primary" id="request-player-one">Dibs!</button>');
                } else if (data.PlayerOne == Quarto.socket().getUsername()) {
                	playerOne.html('<button type="button" class="btn btn-danger" id="leave-player-one">You, Leave?</button>')
                } else {
                	playerOne.html('<span class="label label-default">' + data.PlayerOne + '</span>');
                }
                if (!data.PlayerTwo) {
                	playerTwo.html('<button type="button" class="btn btn-primary" id="request-player-two">Dibs!</button>');
                } else if (data.PlayerTwo == Quarto.socket().getUsername()) {
                	playerTwo.html('<button type="button" class="btn btn-danger" id="leave-player-two">You, Leave?</button>')
                } else {
                	playerTwo.html('<span class="label label-default">' + data.PlayerTwo + '</span>');
                }
                $('#privacy').text(data.Private ? "Private" : "Public");
                observers.empty();
                $(data.Observers).each(function(index, observer) {
                	observers.append('<li><span class="label label-default">' + observer + '</span></li>')
                });
            });

			$('#game-div').on('click', '#request-player-one', function () {
				Quarto.socket().sendMessage(Quarto.constants.GamePlayerOneRequest);
			});

			$('#game-div').on('click', '#request-player-two', function () {
				Quarto.socket().sendMessage(Quarto.constants.GamePlayerTwoRequest);
			});

			$('#game-div').on('click', '#leave-player-one', function () {
				Quarto.socket().sendMessage(Quarto.constants.GamePlayerOneLeave);
			});

			$('#game-div').on('click', '#leave-player-two', function () {
				Quarto.socket().sendMessage(Quarto.constants.GamePlayerTwoLeave);
			});

			$('#leave-room').on('click', function () {
				Quarto.main().loadWaitingRoomHTML();
			});
		}

		function stop() {
			roomName = undefined;
			playerOne = undefined;
			playerTwo = undefined;
			private = undefined;
			observers = undefined;
			$(document).off(Quarto.constants.RoomChange);
			$('#game-div').off('click');
			$('#leave-room').off('click');
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