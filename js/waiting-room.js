(function ($) {

	var userRowConstant = "<tr data-user-uuid='{UserUuid}'><td><div>{Username}</div></td><td><div>{Challenge}</div></td><td><div>{RoomName}</div></td></tr>"
	var roomRowConstant = "<tr data-room-urid='{RoomUrid}'><td><div>{Name}<div></td><td><div>{Members}<div></td><td><div>{Private}<div></td><td><div>{Join}<div></td></tr>"
	
	Quarto.waitingRoom = (function () {

		function loadWaitingRoom() {
			$('body').load('views/waiting-room.html', function() {
				$.get("/users", function (data) {
					console.log(data);
					var usersTable = $('#users-table');
					$(data).each(function (index, user) {
						var roomName = user.RoomName == "" ? "None" : user.RoomName;
						var newRow = $(userRowConstant.replace("{UserUuid}", user.Uuid)
												.replace("{Username}", user.Username)
												.replace("{Challenge}", user.Uuid)
												.replace("{RoomName}", roomName));
						usersTable.append(newRow);
					});
				}, "json");
				$.get("/rooms", function (data) {
					console.log(data);
					var roomsTable = $('#rooms-table');
					$(data).each(function (index, room) {
						var newRow = $(roomRowConstant.replace("{RoomUrid}", room.Urid)
												.replace("{Name}", room.Name)
												.replace("{Members}", room.Members)
												.replace("{Private}", room.Private)
												.replace("{Join}", "<a href='#' class='join-room' data-room-uuid='" + room.Urid + "'>Join</a>"));
						roomsTable.append(newRow);
					});


				$('#rooms-table').on('click', '.join-room', function (event) {
					console.log("Joining " + $(this).data('room-uuid'));

					var room = JSON.stringify({
						Action: Quarto.constants.joinedRoom,
						Data: JSON.stringify({
							Urid: $(this).data('room-uuid')
						})
					});
					Quarto.socket().sendMessage("server", room);

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
					console.log("Submitting room for " + $('#new-room-name').val());
					$('#create-room').hide('slow');
					$('#create-new-room').removeClass('disabled');

					var room = JSON.stringify({
						Action: Quarto.constants.addRoom,
						Data: JSON.stringify({
							Name: $('#new-room-name').val(),
							Private: false,
							Password: ""
						})
					});

					Quarto.socket().sendMessage("server", room);
					Quarto.main().loadGameHTML();
				});
			});
		}

		return {
			loadWaitingRoom: loadWaitingRoom
		}
	});

	$(document).on(Quarto.constants.addRoom, function(event, data) {
		var room = JSON.parse(data.Data);
		var newRow = $(roomRowConstant.replace("{RoomUuid}", room.Urid)
								.replace("{Name}", room.Name)
								.replace("{Members}", room.Members)
								.replace("{Private}", room.Private)
								.replace("{Join}", "<a href='#' class='join-room' data-room-uuid='" + room.Urid + "'>Join</a>"));
		$('#rooms-table').append(newRow);
		newRow.find('div').hide().slideDown();
	});

	$(document).on(Quarto.constants.changeRoom, function(event, data) {
		var room = JSON.parse(data.Data);
		var changedRow = $('#rooms-table [data-room-urid=' + room.Urid + ']');

		changedRow.empty().html(roomRowConstant.replace("{RoomUuid}", room.Urid)
								.replace("{Name}", room.Name)
								.replace("{Members}", room.Members)
								.replace("{Private}", room.Private)
								.replace("{Join}", "<a href='#' class='join-room' data-room-uuid='" + room.Urid + "'>Join</a>"));
		$(changedRow).animateHighlight("#dd0000", 1000);
	});

	$(document).on(Quarto.constants.removeRoom, function(event, data) {
		var room = JSON.parse(data.Data);
		var changedRow = $('#rooms-table [data-room-urid=' + room.Urid + ']');
		changedRow.slideUp(function() {
			changedRow.remove();
		});
	});

	$(document).on(Quarto.constants.addUser, function(event, data) {
		var user = JSON.parse(data.Data);
		if (user.Uuid == Quarto.socket().getUuid()) return;
		var roomName = user.RoomName == "" ? "None" : user.RoomName;
		var newRow = $(userRowConstant.replace("{UserUuid}", user.Uuid)
										.replace("{Username}", user.Username)
										.replace("{Challenge}", user.Uuid)
										.replace("{RoomName}", roomName));
		$('#users-table').append(newRow);
		newRow.find('div').hide().slideDown();
	});

	$(document).on(Quarto.constants.removeUser, function(event, data) {
		var user = JSON.parse(data.Data);
		var changedRow = $('#users-table [data-user-uuid=' + user.Uuid + ']');
		changedRow.find('div').slideUp(function() {
			changedRow.remove();
		});
	});

})(jQuery);