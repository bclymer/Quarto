(function ($) {

	var userRowConstant = "<tr data-user-name='{Username}'><td><div>{Username}</div></td><td><div>{Challenge}</div></td><td><div>{RoomName}</div></td></tr>"
	var roomRowConstant = "<tr data-room-name='{RoomName}'><td><div>{Name}<div></td><td><div>{Members}<div></td><td><div>{Private}<div></td><td><div>{Join}<div></td></tr>"
	
	Quarto.waitingRoom = (function () {

		function loadWaitingRoom() {
			$('body').load('views/waiting-room.html', function() {
				$.get("/users", function (data) {
					console.log(data);
					var usersTable = $('#users-table');
					$(data).each(function (index, user) {
						var roomName = user.RoomName == "" ? "None" : user.RoomName;
						var newRow = $(userRowConstant.replace("{Username}", user.Username)
												.replace("{Username}", user.Username)
												.replace("{Challenge}", user.Username)
												.replace("{RoomName}", roomName));
						usersTable.append(newRow);
					});
				}, "json");
				$.get("/rooms", function (data) {
					console.log(data);
					var roomsTable = $('#rooms-table');
					$(data).each(function (index, room) {
						var newRow = $(roomRowConstant.replace("{RoomName}", room.Name)
												.replace("{Name}", room.Name)
												.replace("{Members}", room.Members)
												.replace("{Private}", room.Private)
												.replace("{Join}", "<a href='#' class='join-room' data-room-name='" + room.Name + "'>Join</a>"));
						roomsTable.append(newRow);
					});


				$('#rooms-table').on('click', '.join-room', function (event) {
					console.log("Joining " + $(this).data('room-name'));

					var room = JSON.stringify({
						Action: Quarto.constants.joinedRoom,
						Data: JSON.stringify({
							Urid: $(this).data('room-name')
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

	$(document).on(Quarto.constants.addRoom, function(event, room) {
		var newRow = $(roomRowConstant.replace("{RoomName}", room.Name)
								.replace("{Name}", room.Name)
								.replace("{Members}", room.Members)
								.replace("{Private}", room.Private)
								.replace("{Join}", "<a href='#' class='join-room' data-room-name='" + room.Name + "'>Join</a>"));
		$('#rooms-table').append(newRow);
		newRow.find('div').hide().slideDown();
	});

	$(document).on(Quarto.constants.changeRoom, function(event, room) {
		var changedRow = $('#rooms-table [data-room-name=' + room.Name + ']');

		changedRow.empty().html(roomRowConstant.replace("{RoomName}", room.Name)
								.replace("{Name}", room.Name)
								.replace("{Members}", room.Members)
								.replace("{Private}", room.Private)
								.replace("{Join}", "<a href='#' class='join-room' data-room-name='" + room.Name + "'>Join</a>"));
		$(changedRow).animateHighlight("#dd0000", 1000);
	});

	$(document).on(Quarto.constants.removeRoom, function(event, user) {
		var changedRow = $('#rooms-table [data-room-name="' + room.Name + '"]');
		changedRow.slideUp(function() {
			changedRow.remove();
		});
	});

	$(document).on(Quarto.constants.addUser, function(event, user) {
		if ($('#users-table [data-user-name="' + user.Username + '"]').length > 0) return;

		var newRow = $(userRowConstant.replace("{Username}", user.Username)
										.replace("{Username}", user.Username)
										.replace("{Challenge}", user.Username)
										.replace("{RoomName}", "None"));
		$('#users-table').append(newRow);
		newRow.find('div').hide().slideDown();
	});

	$(document).on(Quarto.constants.removeUser, function(event, user) {
		var changedRow = $('#users-table [data-user-name="' + user.Username + '"]');
		changedRow.find('div').slideUp(function() {
			changedRow.remove();
		});
	});

})(jQuery);