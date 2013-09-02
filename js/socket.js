(function ($) {

	var socket;
	var uuid;
	var username;
	
	Quarto.socket = (function () {

		function makeConnection(requestedUsername) {
			username = requestedUsername;
			socket = new WebSocket('ws://' + window.location.host + '/realtime?username=' + username);
			socket.onmessage = function (event) {
				console.log(event.data);
				if (!event.data) {
					return;
				}
				var message = JSON.parse(event.data);
				if (!message.Data) {
					return;
				}
				data = JSON.parse(message.Data);
				console.log("Triggering " + message.Action)
				$(document).trigger(message.Action, data);
			};
		}

		function sendMessage(action, message) {
			socket.send(JSON.stringify({
				Action: action,
				Data: message
			}));
		}

		function getUuid() {
			return uuid;
		}

		function getUsername() {
			return username;
		}

		return {
			makeConnection: makeConnection,
			sendMessage: sendMessage,
			getUuid: getUuid,
			getUsername: getUsername,
		}

	});

	$(document).on(Quarto.constants.uuidAssigned, function (event, data) {
		console.log("Uuid for session set to " + data.Data);
		uuid = data.Data;
	});


})(jQuery);