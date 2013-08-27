(function ($) {

	var uuid;
	var opponent;

	var socket;
	
	Quarto.socket = (function () {

		function makeConnection(requestedUuid) {
			uuid = requestedUuid;
			socket = new WebSocket('ws://' + window.location.host + '/realtime?uuid=' + uuid);
			socket.onmessage = function (event) {
				var message = JSON.parse(event.data);
				if (message.Uuid == uuid || (message.Uuid != opponent && opponent)) {
					console.log("Rejecting message, not opponent.");
					return;
				}

				if (message.Action == "action") {
					data = JSON.parse(message.Data);
					data.Uuid = message.Uuid;
					$(document).trigger(data.action, data);
				} else {
					$(document).trigger(message.Action, message);
				}
			};
		}

		function sendMessage(action, message) {
			if (!opponent) return;

			socket.send(JSON.stringify({
				action: action,
				data: message,
				ToUuid: opponent
			}));
		}

		return {
			makeConnection: makeConnection,
			sendMessage: sendMessage
		}

	});

	function recieveMessage(event) {
		var message = JSON.parse(event.data);
		if (message.Uuid == uuid || (message.Uuid != opponent && opponent)) {
			console.log("Rejecting message, not opponent.");
			return;
		}

		if (message.Action == "action") {
			data = JSON.parse(message.Data);
			data.Uuid = message.Uuid;
			$(document).trigger(data.action, data);
		} else {
			$(document).trigger(message.Action, message);
		}
	}

	$(document).on("joined", function (event, data) {
		opponent = data.Uuid;
		Quarto.socket().sendMessage("accept", uuid);
		console.log("send: accepted partner");
	});

	$(document).on("accept", function (event, data) {
		opponent = data.data;
		console.log("accpeted partner: " + opponent);
	});

	$(document).on("left", function (event, data) {
		console.log("player left");
		opponent = undefined;
	});


})(jQuery);