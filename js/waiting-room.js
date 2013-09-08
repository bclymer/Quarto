(function ($) {

	var userRowTemplate = Handlebars.compile($("#user-template").html());
	var roomRowTemplate = Handlebars.compile($("#room-template").html());
	var loaded = false;
	
	Quarto.waitingRoom = (function () {

		function start() {
			loaded = true;
			console.log("waitingRoom start()");
			$('#waiting-room').show();
			$.get("/users", function (data) {
				console.log(data);
				var usersTable = $('#users-table');
				$('#users-table .non-head').remove();
				$(data).each(function (index, user) {
					if ($('#users-table [data-user-name="' + user.Username + '"]').length > 0) return;

					usersTable.append(userRowTemplate(user));
				});
			}, "json");
			$.get("/rooms", function (data) {
				console.log(data);
				var roomsTable = $('#rooms-table');
				$('#rooms-table .non-head').remove();
				$(data).each(function (index, room) {
					if ($('#rooms-table [data-room-name="' + room.Name + '"]').length > 0) return;

					roomsTable.append(roomRowTemplate(room));
				});

				$('#rooms-table').on('click', '.join-room', function (event) {
					var room = JSON.stringify({
						Name: $(this).data('room-name')
					})
					Quarto.socket().sendMessage(Quarto.constants.UserRoomJoin, room);
					Quarto.main().loadGameHTML();

					event.preventDefault();
				});

			}, "json");

			$('#create-new-room').on('click', function (event) {
				$('#create-room').show(function() {
					$('#new-room-name').focus();
				});
				$('#create-new-room').addClass('disabled');
			});

			$('#submit-new-room').on('click', function (event) {
				$('#create-room').hide('slow');
				$('#create-new-room').removeClass('disabled');

				var room = JSON.stringify({
					Name: $('#new-room-name').val(),
					Private: false,
					Password: ""
				});
				$('#new-room-name').val("");
				Quarto.socket().sendMessage(Quarto.constants.RoomAdd, room);
				Quarto.main().loadGameHTML();
			});

			$(document).on(Quarto.constants.RoomAdd, function(event, room) {
				if ($('#rooms-table [data-room-name="' + room.Name + '"]').length > 0) return;

				$('#rooms-table').append(roomRowTemplate(room));
			});

			$(document).on(Quarto.constants.RoomChange, function(event, room) {
				var changedRow = $('#rooms-table [data-room-name="' + room.Name + '"]');
				var cells = changedRow.find('td');
				$(cells[0]).text(room.Name);
				$(cells[1]).text(room.Members);
				$(cells[2]).text(room.Private ? "Yes" : "No");
				$(cells[3]).html('<a href="#" class="join-room" data-room-name="' + room.Name + '">Join</a>');
			});

			$(document).on(Quarto.constants.RoomRemove, function(event, room) {
				var changedRow = $('#rooms-table [data-room-name="' + room.Name + '"]');
				changedRow.slideUp(function() {
					changedRow.remove();
				});
			});

			$(document).on(Quarto.constants.UserRoomJoin, function(event, userRoom) {
				var changedRow = $('#users-table [data-user-name="' + userRoom.Username + '"]');
				$(changedRow.find('td')[2]).html(userRoom.RoomName);
		        changedRow.animate({backgroundColor: "#dd0000"}, 500, function () {
		        	changedRow.animate({backgroundColor: ""}, 500)
		        });
			});

			$(document).on(Quarto.constants.UserAdd, function(event, user) {
				if ($('#users-table [data-user-name="' + user.Username + '"]').length > 0) return;

				$('#users-table').append(userRowTemplate(user));
			});

			$(document).on(Quarto.constants.UserRemove, function(event, user) {
				var changedRow = $('#users-table [data-user-name="' + user.Username + '"]');
				changedRow.find('div').slideUp(function() {
					changedRow.remove();
				});
			});
		}

		function stop() {
			$('#submit-new-room').off('click');
			$('#rooms-table').off('click');
			$('#create-new-room').off('click');
			$('#waiting-room').hide();
			$(document).off(Quarto.constants.RoomAdd)
						.off(Quarto.constants.RoomChange)
						.off(Quarto.constants.RoomRemove)
						.off(Quarto.constants.UserRoomJoin)
						.off(Quarto.constants.UserAdd)
						.off(Quarto.constants.UserRemove);
			loaded = false;
			console.log("waitingRoom stop()");
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

})(jQuery);