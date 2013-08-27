(function ($) {

	var chatTemplate = '<div class="message-container {2}"><span class="sender">{0}</span><span class="message"> - {1}</span></div>';
	var opponent;
	var sendButton;
	var chatInput;
	var chatDiv;
	
	Quarto.chat = (function() {

		function attachEventsToGamePage() {
			chatInput = $('#chat');
			sendButton = $('#send');
			chatDiv = $('#chat-div');

			sendButton.click(function () {
				Quarto.socket().sendMessage("chat", chatInput.val());
				chatDiv.append(chatTemplate.replace("{0}", "You").replace("{1}", chatInput.val()).replace("{2}", "you"));
				scrollChatToBottom();
				chatInput.val("");
			});

			chatInput.keydown(function (e) {
				if (e.which == 13 && chatInput.val().length > 0 && opponent) {
					sendButton.click();
					return false;
				}
			});

			chatInput.keyup(function (e) {
				if (chatInput.val().length > 0 && opponent) {
					sendButton.removeClass("disabled");
				} else {
					sendButton.addClass("disabled");
				}
			});
		}

		return {
			attachEventsToGamePage: attachEventsToGamePage
		}

	});

	$(document).on("keydown", function (event) {
		chatInput.focus();
	});

	$(document).on("chat", function (event, data) {
		chatDiv.append(chatTemplate.replace("{0}", opponent).replace("{1}", data.data).replace("{2}", "opponent"));
		scrollChatToBottom();
	});

	$(document).on("left", function (event, data) {
		chatDiv.empty();
	});

	function scrollChatToBottom() {
		var chatDiv = document.getElementById("chat-div");
		chatDiv.scrollTop = chatDiv.scrollHeight;
	}

	$(document).on("accept", function (event, data) {
		chatDiv.empty();
		chatDiv.append(chatTemplate.replace("{0}", "System").replace("{1}", data.data + " has joined your game.").replace("{2}", "opponent"));
	});

	$(document).on("joined", function (event, data) {
		chatDiv.empty();
		chatDiv.append(chatTemplate.replace("{0}", "System").replace("{1}", data.Uuid + " has joined your game.").replace("{2}", "opponent"));
	});

	$(document).on("joined", function (event, data) {
		opponent = data.Uuid;
	});

	$(document).on("accept", function (event, data) {
		opponent = data.data;
	});

	$(document).on("left", function (event, data) {
		opponent = undefined;
	});

})(jQuery);