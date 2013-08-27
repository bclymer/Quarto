(function ($) {

	var chatTemplate = '<div class="message-container {2}"><span class="sender">{0}</span><span class="message"> - {1}</span></div>';
	var opponent;
	
	Quarto.chat = (function() {

		function attachEventsToGamePage() {
			$('#game-div').hide();
			$('#game-div').fadeIn(500);
			var chatInput = $('#chat');

			$('#send').click(function () {
				Quarto.socket().sendMessage("chat", chatInput.val());
				$('#chat-div').append(chatTemplate.replace("{0}", "You").replace("{1}", chatInput.val()).replace("{2}", "you"));
				scrollChatToBottom();
				chatInput.val("");
			});

			chatInput.keydown(function (e) {
				if (e.which == 13 && chatInput.val().length > 0 && opponent) {
					$('#send').click();
					return false;
				}
			});

			chatInput.keyup(function (e) {
				if (chatInput.val().length > 0 && opponent) {
					$('#send').removeClass('disabled');
				} else {
					$('#send').addClass('disabled');
				}
			});
		}

		return {
			attachEventsToGamePage: attachEventsToGamePage
		}

	});

	$(document).on("keydown", function (event) {
		$("#chat").focus();
	});

	$(document).on("chat", function (event, data) {
		$('#chat-div').append(chatTemplate.replace("{0}", opponent).replace("{1}", data.data).replace("{2}", "opponent"));
		scrollChatToBottom();
	});

	$(document).on("left", function (event, data) {
		$('#chat-div').empty();
	});

	function scrollChatToBottom() {
		var chatDiv = document.getElementById("chat-div");
		chatDiv.scrollTop = chatDiv.scrollHeight;
	}

	$(document).on("accept", function (event, data) {
		$('#chat-div').empty();
		$('#chat-div').append(chatTemplate.replace("{0}", "System").replace("{1}", data.data + " has joined your game.").replace("{2}", "opponent"));
	});

	$(document).on("joined", function (event, data) {
		$('#chat-div').empty();
		$('#chat-div').append(chatTemplate.replace("{0}", "System").replace("{1}", data.Uuid + " has joined your game.").replace("{2}", "opponent"));
	});

	$(document).on("joined", function (event, data) {
		opponent = data.Uuid;
		console.log("chat opponent set to " + opponent);
	});

	$(document).on("accept", function (event, data) {
		opponent = data.data;
		console.log("chat opponent set to " + opponent);
	});

	$(document).on("left", function (event, data) {
		opponent = undefined;
		console.log("chat opponent set to " + opponent);
	});

})(jQuery);