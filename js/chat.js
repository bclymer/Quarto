(function ($) {

	var chatTemplate = Handlebars.compile($("#chat-template").html());
	var sendButton;
	var chatInput;
	var chatDiv;
	var cachedEvents = [];
	var loaded = false;
	
	Quarto.chat = (function() {

		function start() {
			loaded = true;
			console.log("chat start()");
			chatInput = $('#chat');
			sendButton = $('#send');
			chatDiv = $('#chat-div');

			sendButton.on('click', function () {
				var chatData = JSON.stringify({
					Message: chatInput.val()
				});
				Quarto.socket().sendMessage(Quarto.constants.Chat, chatData);
				scrollChatToBottom();
				chatInput.val("");
			});

			chatInput.keydown(function (e) {
				if (e.which == 13 && chatInput.val().length > 0) {
					sendButton.click();
					return false;
				}
			});

			chatInput.on('keyup', function (e) {
				if (chatInput.val().length > 0) {
					sendButton.removeClass("disabled");
				} else {
					sendButton.addClass("disabled");
				}
			});
			
			$.each(cachedEvents, function (index, data) {
				applyMessage(data);
			});
			cachedEvents = [];

			$(document).on('keydown', function(event) {
				if (chatInput) {
					chatInput.focus();
				}
			});

			$(document).on(Quarto.constants.Chat, function (event, data) {
				if (!chatDiv) {
					cachedEvents.push(data)
					return;
				} else {
					applyMessage(data);
				}
			});

			$(document).on(Quarto.constants.UserRoomJoin, function (event, data) {
				if (!chatDiv) return;
				var message = {
					Sender: "opponent",
					Username: "System",
					Message: data.Username + " joined the room."
				}
				chatDiv.append(chatTemplate(message));
			});

			$(document).on(Quarto.constants.UserRoomLeave, function (event, data) {
				if (!chatDiv) return;
				var message = {
					Sender: "opponent",
					Username: "System",
					Message: data.Username + " left the room."
				}
				chatDiv.append(chatTemplate(message));
			});
		}

		function stop() {
			chatInput.off();
			sendButton.off();
			chatDiv.off();
			$(document).off('keydown');
			$(document).off(Quarto.constants.Chat);
			$(document).off(Quarto.constants.UserRoomJoin);
			$(document).off(Quarto.constants.UserRoomLeave);

			cachedEvents = [];
			chatInput = undefined;
			sendButton = undefined;
			chatDiv = undefined;
			loaded = false;
			console.log("chat stop()");
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

	function applyMessage(data) {
		if (!chatDiv) return;
		data.Sender = (data.Username == Quarto.socket().getUsername()) ? "you" : "opponent";
		chatDiv.append(chatTemplate(data));
		scrollChatToBottom();
	}

	function scrollChatToBottom() {
		if (!chatDiv) return;
		var chatDiv = document.getElementById("chat-div");
		chatDiv.scrollTop = chatDiv.scrollHeight;
	}

})(jQuery);