(function ($) {

	var chatTemplate = '<div class="message-container {2}"><span class="sender">{0}</span><span class="message"> - {1}</span></div>';
	var sendButton;
	var chatInput;
	var chatDiv;
	var cachedEvents = [];
	
	Quarto.chat = (function() {

		function attachEventsToGamePage() {
			chatInput = $('#chat');
			sendButton = $('#send');
			chatDiv = $('#chat-div');

			sendButton.click(function () {
				var chatData = JSON.stringify({
					username: Quarto.socket().getUsername(),
					message: chatInput.val(),
					uuid: Quarto.socket().getUuid(),
				});
				Quarto.socket().sendMessage(Quarto.constants.chat, chatData);
				scrollChatToBottom();
				chatInput.val("");
			});

			chatInput.keydown(function (e) {
				if (e.which == 13 && chatInput.val().length > 0) {
					sendButton.click();
					return false;
				}
			});

			chatInput.keyup(function (e) {
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
		}

		return {
			attachEventsToGamePage: attachEventsToGamePage
		}

	});

	$(document).on("keydown", function (event) {
		if (chatInput) {
			chatInput.focus();
		}
	});

	$(document).on(Quarto.constants.chat, function (event, data) {
		if (!chatDiv) {
			cachedEvents.push(data)
			return;
		} else {
			applyMessage(data);
		}
	});

	function applyMessage(data) {
		var chatData = JSON.parse(data.Data);
		if (chatData.uuid == Quarto.socket().getUuid()) {
			chatDiv.append(chatTemplate.replace("{0}", "You").replace("{1}", chatData.message).replace("{2}", "you"));
		} else {
			chatDiv.append(chatTemplate.replace("{0}", chatData.username).replace("{1}", chatData.message).replace("{2}", "opponent"));
		}
		scrollChatToBottom();
	}

	function scrollChatToBottom() {
		var chatDiv = document.getElementById("chat-div");
		chatDiv.scrollTop = chatDiv.scrollHeight;
	}

	$(document).on(Quarto.constants.joinedRoom, function (event, data) {
		if (!chatDiv) return;
		chatDiv.append(chatTemplate.replace("{0}", "System").replace("{1}", data.Data + " joined the room.").replace("{2}", "opponent"));
	});

	$(document).on(Quarto.constants.leftRoom, function (event, data) {
		if (!chatDiv) return;
		chatDiv.append(chatTemplate.replace("{0}", "System").replace("{1}", data.Data + " left the room.").replace("{2}", "opponent"));
	});

})(jQuery);