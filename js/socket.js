(function ($) {

	var socket;
	var username;
	
	Quarto.socket = (function () {

		function makeConnection(requestedUsername, onOpen) {
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
			socket.onopen = onOpen();
			socket.onclose = function () {
				Quarto.main().loadRegisterHTML();
			};
			socket.onerror = function() {
				toastr.error("Lost Connection to Server");
				socket.close();
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

		function disconnect() {
			socket.close();
			document.cookie = undefined;
		}

		return {
			makeConnection: makeConnection,
			sendMessage: sendMessage,
			getUsername: getUsername,
		}

	});


})(jQuery);