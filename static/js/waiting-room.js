(function ($) {

	var userRowTemplate = Handlebars.compile($("#user-template").html());
	var roomRowTemplate = Handlebars.compile($("#room-template").html());
	var loaded = false;
	
	Quarto.waitingRoom = (function () {

		function start() {
			loaded = true;
			console.log("waitingRoom start()");
			$('#waiting-room').show();
			$('#profile-dropdown').text(Quarto.socket().getUsername());
			$.get("/users", function (data) {
				console.log(data);
				var usersTable = $('#users-table');
				$('#users-table .non-head').remove();
				$(data).each(function (index, user) {
					if ($('#users-table [data-user-name="' + user.username + '"]').length > 0) return;

					usersTable.append(userRowTemplate(user));
				});
			}, "json");
			$.get("/rooms", function (data) {
				console.log(data);
				var roomsTable = $('#rooms-table');
				$('#rooms-table .non-head').remove();
				$(data).each(function (index, room) {
					if ($('#rooms-table [data-room-name="' + room.name + '"]').length > 0) return;

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
				$('#create-room').bPopup({
					onOpen: function() {
						$('#new-room-name').on('keydown', function (e) {
							if (e.which == 13 && $('#new-room-name').val().length > 0) {
								$('#submit-new-room').click();
								return false;
							}
						});
					},
					onClose: function() {
						$('#new-room-name').off('keydown');
					}
				});
			});

			$('#submit-new-room').on('click', function (event) {
				$('#create-room').bPopup().close();
				$('#create-room').hide('slow');
				$('#create-new-room').removeClass('disabled');

				var room = JSON.stringify({
					name: $('#new-room-name').val(),
					private: false,
					password: ""
				});
				$('#new-room-name').val("");
				Quarto.socket().sendMessage(Quarto.constants.RoomAdd, room);
				Quarto.main().loadGameHTML();
			});

			$('#logout').on('click', function () {
				console.log("Logout!");
			});

			$(document).on(Quarto.constants.RoomAdd, function(event, room) {
				if ($('#rooms-table [data-room-name="' + room.name + '"]').length > 0) return;

				$('#rooms-table').append(roomRowTemplate(room));
			});

			$(document).on(Quarto.constants.RoomChange, function(event, room) {
				var changedRow = $('#rooms-table [data-room-name="' + room.name + '"]');
				var cells = changedRow.find('td');
				$(cells[0]).text(room.name);
				$(cells[1]).text(room.members);
				$(cells[2]).text(room.private ? "Yes" : "No");
				$(cells[3]).html('<a href="#" class="join-room" data-room-name="' + room.name + '">Join</a>');
			});

			$(document).on(Quarto.constants.RoomRemove, function(event, room) {
				var changedRow = $('#rooms-table [data-room-name="' + room.name + '"]');
				changedRow.slideUp(function() {
					changedRow.remove();
				});
			});

			$(document).on(Quarto.constants.UserRoomJoin, function(event, userRoom) {
				var changedRow = $('#users-table [data-user-name="' + userRoom.username + '"]');
				$(changedRow.find('td')[2]).html(userRoom.roomName);
		        changedRow.animate({backgroundColor: "#dd0000"}, 500, function () {
		        	changedRow.animate({backgroundColor: ""}, 500)
		        });
			});

			$(document).on(Quarto.constants.UserAdd, function(event, user) {
				if ($('#users-table [data-user-name="' + user.username + '"]').length > 0) return;

				$('#users-table').append(userRowTemplate(user));
			});

			$(document).on(Quarto.constants.UserRemove, function(event, user) {
				var changedRow = $('#users-table [data-user-name="' + user.username + '"]');
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