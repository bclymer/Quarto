(function ($) {

	var socket;
	var username;
	
	Quarto.socket = (function () {

		function makeConnection(requestedUsername, onOpen) {
			username = requestedUsername;
			socket = new WebSocket('ws://' + window.location.host + '/quarto/realtime?username=' + username);
			socket.onmessage = function (event) {
				console.log("In: " + event.data);
				if (!event.data) {
					return;
				}
				var message = JSON.parse(event.data);
				if (!message.data) {
					return;
				}
				data = JSON.parse(message.data);
				$(document).trigger(message.action, data);
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
				action: action,
				data: message
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