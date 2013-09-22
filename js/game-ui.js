(function ($) {

	var loaded = false;
	var gameState;
	var roomName;
	var playerOne;
	var playerTwo;
	var private;
	var observers;
	
	Quarto.gameUi = (function () {

		function start() {
			loaded = true;
			console.log("gameUi start()");
			gameState = $('#game-state');
			roomName = $('#room-name');
			playerOne = $('#player-one');
			playerTwo = $('#player-two');
			private = $('#private');
			observers = $('#observers');

            $(document).on(Quarto.constants.RoomChange, function (event, data) {
                roomName.text(data.name);
                if (!data.playerOne) {
                	playerOne.html('<button type="button" class="btn btn-primary" id="request-player-one">Dibs!</button>');
                } else if (data.playerOne == Quarto.socket().getUsername()) {
                	playerOne.html('<button type="button" class="btn btn-danger" id="leave-player-one">You, Leave?</button>')
                } else {
                	playerOne.html('<span class="label label-default">' + _.escape(data.playerOne) + '</span>');
                }
                if (!data.playerTwo) {
                	playerTwo.html('<button type="button" class="btn btn-primary" id="request-player-two">Dibs!</button>');
                } else if (data.playerTwo == Quarto.socket().getUsername()) {
                	playerTwo.html('<button type="button" class="btn btn-danger" id="leave-player-two">You, Leave?</button>')
                } else {
                	playerTwo.html('<span class="label label-default">' + _.escape(data.playerTwo) + '</span>');
                }
                $('#privacy').text(data.private ? "Private" : "Public");
                observers.empty();
                $(data.observers).each(function(index, observer) {
                	observers.append('<li><span class="label label-default">' + _.escape(observer) + '</span></li>')
                });
            });

			$(document).on(Quarto.constants.GameWinner, function (event, data) {
				$('#dialog-message').text(data.winner + " won!");
				$('#popup').bPopup();
			});

			$(document).on(Quarto.constants.GameChange, function (event, data) {
				var gameStateText;
            	switch(data.gameState) {
            		case 0:
            			gameStateText = "Waiting for players";
            			break;
            		case 1:
            			gameStateText = "Player one is choosing a piece.";
            			break;
            		case 2:
            			gameStateText = "Player one is playing the chosen piece.";
            			break;
            		case 3:
            			gameStateText = "Player two is choosing a piece.";
            			break;
            		case 4:
            			gameStateText = "Player two is playing the chosen piece.";
            			break;
            	}
            	gameState.text(gameStateText);
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
			gameState = undefined;
			roomName = undefined;
			playerOne = undefined;
			playerTwo = undefined;
			private = undefined;
			observers = undefined;
			$(document).off(Quarto.constants.RoomChange)
						.off(Quarto.constants.GameWinner);
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