(function ($) {

	var socket;
	var username;
	
	Quarto.socket = (function () {

		function makeConnection(requestedUsername) {
			username = requestedUsername;
			socket = new WebSocket('ws://' + window.location.host + '/realtime?username=' + username);
			socket.onmessage = function (event) {
				console.log("In: " + event.data);
				if (!event.data) {
					return;
				}
				var message = JSON.parse(event.data);
				if (!message.Data) {
					return;
				}
				data = JSON.parse(message.Data);
				$(document).trigger(message.Action, data);
			};
		}

		function sendMessage(action, message) {
			var serializedMessage = JSON.stringify({
				Action: action,
				Data: message
			});
			console.log("Out: " + serializedMessage);
			socket.send(serializedMessage);
		}

		function getUsername() {
			return username;
		}

		return {
			makeConnection: makeConnection,
			sendMessage: sendMessage,
			getUsername: getUsername,
		}

	});


})(jQuery);