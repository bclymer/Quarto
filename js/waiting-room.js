(function ($) {

	var userRowConstant = '<tr class="non-head" data-user-name="{Username}"><td><div>{Username}</div></td><td><div>{Challenge}</div></td><td><div>{RoomName}</div></td></tr>';
	var roomRowConstant = '<tr class="non-head" data-room-name="{RoomName}"><td><div>{Name}<div></td><td><div>{Members}<div></td><td><div>{Private}<div></td><td><div>{Join}<div></td></tr>';

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
					var roomName = user.RoomName == "" ? "None" : user.RoomName;
					var newRow = $(userRowConstant.replace('{Username}', user.Username)
													.replace('{Username}', user.Username)
													.replace('{Challenge}', user.Username)
													.replace('{RoomName}', roomName));
					usersTable.append(newRow);
				});
			}, "json");
			$.get("/rooms", function (data) {
				console.log(data);
				var roomsTable = $('#rooms-table');
				$('#rooms-table .non-head').remove();
				$(data).each(function (index, room) {
					if ($('#rooms-table [data-room-name="' + room.Name + '"]').length > 0) return;
					var newRow = $(roomRowConstant.replace('{RoomName}', room.Name)
													.replace('{Name}', room.Name)
													.replace('{Members}', room.Members)
													.replace('{Private}', room.Private)
													.replace('{Join}', '<a href="#" class="join-room" data-room-name="' + room.Name + '">Join</a>'));
					roomsTable.append(newRow);
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

				Quarto.socket().sendMessage(Quarto.constants.RoomAdd, room);
				Quarto.main().loadGameHTML();
			});

			$(document).on(Quarto.constants.RoomAdd, function(event, room) {
				if ($('#rooms-table [data-room-name="' + room.Name + '"]').length > 0) return;

				var newRow = $(roomRowConstant.replace('{RoomName}', room.Name)
												.replace('{Name}', room.Name)
												.replace('{Members}', room.Members)
												.replace('{Private}', room.Private)
												.replace('{Join}', '<a href="#" class="join-room" data-room-name="' + room.Name + '">Join</a>'));
				$('#rooms-table').append(newRow);
				newRow.find('div').hide().slideDown();
			});

			$(document).on(Quarto.constants.RoomChange, function(event, room) {
				var changedRow = $('#rooms-table [data-room-name="' + room.Name + '"]');

				changedRow.empty().html(roomRowConstant.replace('{RoomName}', room.Name)
														.replace('{Name}', room.Name)
														.replace('{Members}', room.Members)
														.replace('{Private}', room.Private)
														.replace('{Join}', '<a href="#" class="join-room" data-room-name="' + room.Name + '">Join</a>'));
				$(changedRow).animateHighlight("#dd0000", 1000);
			});

			$(document).on(Quarto.constants.RoomRemove, function(event, user) {
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

				var newRow = $(userRowConstant.replace('{Username}', user.Username)
												.replace('{Username}', user.Username)
												.replace('{Challenge}', user.Username)
												.replace('{RoomName}', "None"));
				$('#users-table').append(newRow);
				newRow.find('div').hide().slideDown();
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