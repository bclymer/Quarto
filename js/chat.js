(function ($) {

	var chatTemplate = '<div class="message-container {2}"><span class="sender">{0}</span><span class="message"> - {1}</span></div>';
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
				data.Message = _.escape(data.Message);
				data.Username = _.escape(data.Username);
				if (!chatDiv) {
					cachedEvents.push(data)
					return;
				} else {
					applyMessage(data);
				}
			});

			$(document).on(Quarto.constants.UserRoomJoin, function (event, data) {
				if (!chatDiv) return;
				chatDiv.append(chatTemplate.replace("{0}", "System").replace("{1}", data.Message + " joined the room.").replace("{2}", "opponent"));
			});

			$(document).on(Quarto.constants.UserRoomLeave, function (event, data) {
				if (!chatDiv) return;
				chatDiv.append(chatTemplate.replace("{0}", "System").replace("{1}", data.Message + " left the room.").replace("{2}", "opponent"));
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
		
		if (data.Username == Quarto.socket().getUsername()) {
			chatDiv.append(chatTemplate.replace("{0}", "You").replace("{1}", data.Message).replace("{2}", "you"));
		} else {
			chatDiv.append(chatTemplate.replace("{0}", data.Username).replace("{1}", data.Message).replace("{2}", "opponent"));
		}
		scrollChatToBottom();
	}

	function scrollChatToBottom() {
		if (!chatDiv) return;
		var chatDiv = document.getElementById("chat-div");
		chatDiv.scrollTop = chatDiv.scrollHeight;
	}

})(jQuery);