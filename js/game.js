﻿(function ($) {

    var loaded = false;

    var fullRadius;
    var smallSize = 0.18;
    var farX = 0.8;
    var closeX = 0.266;
    var selectedPiece = -1;
    var context;
    var gameState = 0;

    var game_state_no_player = 0;
    var game_state_waiting_for_piece = 1;
    var game_state_playing_piece = 2;
    var game_state_choosing_piece = 3;
    var game_state_waiting_for_play = 4;

    var shape_circle = 1;
    var shape_square = 2;
    var shape_piece = 3;

    var drawnObjects = [];

    var possiblePieces = [
        {square: true, hole: true, white: true, tall: true},
        {square: true, hole: true, white: true, tall: false},
        {square: true, hole: true, white: false, tall: true},
        {square: true, hole: true, white: false, tall: false},
        {square: true, hole: false, white: true, tall: true},
        {square: true, hole: false, white: true, tall: false},
        {square: true, hole: false, white: false, tall: true},
        {square: true, hole: false, white: false, tall: false},
        {square: false, hole: true, white: true, tall: true},
        {square: false, hole: true, white: true, tall: false},
        {square: false, hole: true, white: false, tall: true},
        {square: false, hole: true, white: false, tall: false},
        {square: false, hole: false, white: true, tall: true},
        {square: false, hole: false, white: true, tall: false},
        {square: false, hole: false, white: false, tall: true},
        { square: false, hole: false, white: false, tall: false }
    ];

    var usedPieces = [];

    var availablePieces = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15];

    var boardLocations = [-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1];

    Quarto.game = (function () {
        function start(uuid) {
            loaded = true;
            console.log("game start()");
            $('#game-div').show();
            $('body').animate({backgroundColor: '#EEEEEE'}, 500);
            Quarto.chat().start();
            Quarto.gameUi().start();
            
            var canvasElement = $('canvas');
            context = canvasElement.get(0).getContext('2d');
            context.canvas.addEventListener('mousedown', mouseClick, false);

            resetState();

            context.canvas.width = window.innerWidth - 600;
            context.canvas.height = window.innerHeight;
            draw();

            function mouseClick(event) {
                for (var i = 0; i < drawnObjects.length; i++) {
                    var object = drawnObjects[i];
                    if (object.containsPoint(event.x - 300, event.y)) {
                        if (object.onClick) {
                            object.onClick(object);
                        }
                    }
                }
            }

            $(document).on('accept', function (event, data) {
                gameState = game_state_choosing_piece;
                resetState();
                draw();
            });

            $(document).on('chosen', function (event, data) {
                var pieceId = parseInt(data.data);
                selectedPiece = pieceId;

                var piece;
                for (var i = 0; i < drawnObjects.length; i++) {
                    if (drawnObjects[i].shape == 3 && drawnObjects[i].data.pieceId == pieceId) {
                        piece = drawnObjects[i];
                        break;
                    }
                }

                gameState = game_state_playing_piece;
                pieceChosen(piece);
            });

            $(document).on('placed', function (event, data) {
                gameState = game_state_waiting_for_piece;
                locationChosen(data.data);
            });

            $(document).on(Quarto.constants.gameChange, function (event, data) {

            });

            $(window).on('resize', function () {
                context.canvas.width = window.innerWidth - 600;
                context.canvas.height = window.innerHeight;
                draw();
            });
        }

        function stop() {
            $(window).off('resize');
            $('#game-div').hide();
            $(document).off(Quarto.constants.userJoinRoom)
                        .off(Quarto.constants.userLeaveRoom);
            Quarto.socket().sendMessage(Quarto.constants.userLeaveRoom, "");
            loaded = false;
            console.log("game stop()");
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

    // 1: circle, 2: square, 3: piece
    function DrawnObject(x, y, shape, size, data, onClick) {
        this.x = x;
        this.y = y;
        this.shape = shape;
        this.size = size;
        this.data = data;
        this.onClick = onClick;
        if (shape == 3) {
            this.size = data.square ? 0.22 : 0.11;
        }
    }

    DrawnObject.prototype.getRelativeCoordinates = function () {
        return [this.x, this.y];
    }

    DrawnObject.prototype.getAbsoluteCoordinates = function () {
        return [context.canvas.width / 2 + (fullRadius * this.x), context.canvas.height / 2 + (fullRadius * this.y)]
    }

    DrawnObject.prototype.containsPoint = function (x, y) {
        var coordinates = this.getAbsoluteCoordinates();
        var size = this.size;
        if (this.shape == 1) {
            return inCircle();
        } else if (this.shape == 2) {
            return inSquare();
        } else if (this.shape == 3) {
            if (this.data.square) {
                return inSquare();
            } else {
                return inCircle();
            }
        }
        return false;

        function inCircle() {
            return (x > coordinates[0] - size * fullRadius &&
                x < coordinates[0] + size * fullRadius &&
                y > coordinates[1] - size * fullRadius &&
                y < coordinates[1] + size * fullRadius &&
                Math.sqrt(Math.pow(x - coordinates[0], 2) + Math.pow(y - coordinates[1], 2)) < size * fullRadius);
        }

        function inSquare() {
            return (x > coordinates[0] - size * fullRadius / 2 &&
                x < coordinates[0] + size * fullRadius / 2 &&
                y > coordinates[1] - size * fullRadius / 2 &&
                y < coordinates[1] + size * fullRadius / 2);
        }
    }

    DrawnObject.prototype.draw = function () {
        if (this.shape == 1) {
            drawCircle(this.x, this.y, this.size);
        } else if (this.shape == 2) {
            drawSquare(this.x, this.y, this.size);
        } else if (this.shape == 3) {
            drawPiece([this.x, this.y], this.data.square, this.data.hole, this.data.white, this.data.tall);
        }
    }

    function locationChosen(location) {
        var piece;
        for (var i = 0; i < drawnObjects.length; i++) {
            if (drawnObjects[i].shape == 3 && drawnObjects[i].data.pieceId == selectedPiece) {
                piece = drawnObjects[i];
                break;
            }
        }
        var selectedCoordinates = getLocationXandY(location);
        piece.x = selectedCoordinates[0];
        piece.y = selectedCoordinates[1];
        boardLocations[location] = selectedPiece;
        usedPieces[usedPieces.length] = selectedPiece;
        selectedPiece = -1;
        draw();
        checkForWinner();
    }

    function pieceChosen(piece) {
        var selectedCoordinates = getLocationXandY(16);
        piece.x = selectedCoordinates[0];
        piece.y = selectedCoordinates[1];
        draw();
    }

    function resetState() {
        usedPieces = [];
        drawnObjects = [];
        boardLocations = [-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1];

        drawnObjects[0] = new DrawnObject(0.26, 0, shape_circle, 1);
        for (var i = 0; i < 16; i++) {
            var coordinates = getLocationXandY(i);
            drawnObjects[i + 1] = new DrawnObject(coordinates[0], coordinates[1], shape_circle, smallSize, i,
                function (object) {
                    if (gameState == game_state_playing_piece) {
                        if (boardLocations[object.data] != -1) {
                            return;
                        }
                        gameState = game_state_choosing_piece;
                        locationChosen(object.data);
                        Quarto.socket().sendMessage("placed", object.data);
                    }
                });
        }

        var x = -1.13, y = -0.87, z = 0;
        for (var i = 0; i < availablePieces.length; i++) {
            var pieceId = availablePieces[i];
            var piece = possiblePieces[pieceId];
            drawnObjects[drawnObjects.length] = new DrawnObject(
                x, y, shape_piece, -1 /* overwritten */,
                {
                    square: piece.square,
                    hole: piece.hole,
                    white: piece.white,
                    tall: piece.tall,
                    pieceId: pieceId
                },
                function (object) {
                    if (gameState == game_state_choosing_piece) {
                        if (selectedPiece == -1 && usedPieces.indexOf(object.data.pieceId) == -1) {
                            gameState = game_state_waiting_for_play;
                            pieceChosen(object);
                            selectedPiece = object.data.pieceId;
                            console.log("send: " + selectedPiece);
                            Quarto.socket().sendMessage("chosen", selectedPiece);
                        }
                    }
                }
                );

            y += 0.25;
            if (++z == 8) {
                x += 0.25;
                y = -0.87;
            }
        }
    }

    function draw() {
        context.clearRect(0, 0, context.canvas.width, context.canvas.height);
        fullRadius = Math.min(context.canvas.width / 1.26 / 2, context.canvas.height / 2);

        context.lineWidth = fullRadius / 95.5;
        context.strokeStyle = '#003300';

        for (var i = 0; i < drawnObjects.length; i++) {
            context.beginPath();
            drawnObjects[i].draw();
            context.stroke();
        }
    }

    function checkForWinner() {
        if (checkSequence(function (i, j) { return i * 4 + j; })) {
            return;
        }

        if (checkSequence(function (i, j) { return j * 4 + i; })) {
            return;
        }

        { // limit scope
            var square = 10, hole = 10, white = 10, tall = 10;
            for (var j = 0; j < 16; j += 5) {
                var pieceId = boardLocations[j];
                if (pieceId == -1) {
                    square = 0, hole = 0, white = 0, tall = 0;
                    break;
                }
                var piece = possiblePieces[pieceId];
                square += piece.square;
                hole += piece.hole;
                white += piece.white;
                tall += piece.tall;
            }
            if (checkValues(square, hole, white, tall)) {
                return;
            }
        }

        { // limit scope
            var square = 10, hole = 10, white = 10, tall = 10;
            for (var j = 0; j < 5; j++) {
                var pieceId = boardLocations[j];
                if (pieceId == -1) {
                    square = 0, hole = 0, white = 0, tall = 0;
                    break;
                }
                var piece = possiblePieces[pieceId];
                square += piece.square;
                hole += piece.hole;
                white += piece.white;
                tall += piece.tall;
            }
            if (checkValues(square, hole, white, tall)) {
                return;
            }
        }

        function checkSequence(locationFunction) {
            for (var i = 0; i < 4; i++) {
                var square = 10, hole = 10, white = 10, tall = 10;
                for (var j = 0; j < 4; j++) {
                    var pieceId = boardLocations[locationFunction(i, j)];
                    if (pieceId == -1) {
                        square = 0, hole = 0, white = 0, tall = 0;
                        break;
                    }
                    var piece = possiblePieces[pieceId];
                    square += piece.square;
                    hole += piece.hole;
                    white += piece.white;
                    tall += piece.tall;
                }
                if (checkValues(square, hole, white, tall)) {
                    return true;
                }
            }
            return false;
        }

        function checkValues(square, hole, white, tall) {
            if (isFourOrZero(square) || isFourOrZero(hole) || isFourOrZero(white) || isFourOrZero(tall)) {
                if (gameState == game_state_waiting_for_piece) {
                    $('#dialog-message').text("You Lost!<br />");
                    $("#popup").bPopup();
                } else {
                    $('#dialog-message').text("You Won!<br />");
                    $("#popup").bPopup();
                }
                resetState();
                draw();
                return true;
            }
            return false;
        }
    }

    function isFourOrZero(num) {
        return num == 10 || num == 14;
    }

    function getLocationXandY(location) {
        loc = privateGetLocationXandY(location);
        return [loc[0] + 0.26, loc[1]];
    }

    function privateGetLocationXandY(location) {
        switch (location) {
            case 0:
                return [0, -farX];
            case 1:
                return [closeX, -farX + closeX];
            case 2:
                return [farX - closeX, -closeX];
            case 3:
                return [farX, 0];
            case 4:
                return [-closeX, -farX + closeX];
            case 5:
                return [0, -closeX];
            case 6:
                return [closeX, 0];
            case 7:
                return [farX - closeX, closeX];
            case 8:
                return [-farX + closeX, -closeX];
            case 9:
                return [-closeX, 0];
            case 10:
                return [0, closeX];
            case 11:
                return [closeX, farX - closeX];
            case 12:
                return [-farX, 0];
            case 13:
                return [-farX + closeX, closeX];
            case 14:
                return [-closeX, farX - closeX];
            case 15:
                return [0, farX];
            case 16:
                return [-0.85, -0.85];
        }
    }

    function drawCircle(x, y, size) {
        var radius = fullRadius * size;
        var adjustedX = context.canvas.width / 2 + (fullRadius * x);
        var adjustedY = context.canvas.height / 2 + (fullRadius * y);
        context.moveTo(adjustedX + radius, adjustedY);
        context.arc(adjustedX, adjustedY, radius, 0, 2 * Math.PI, false);
    }

    function drawSquare(x, y, size) {
        var adjustedX = context.canvas.width / 2 + (fullRadius * x);
        var adjustedY = context.canvas.height / 2 + (fullRadius * y);
        var fullSize = size * fullRadius;
        context.moveTo(adjustedX + fullSize / 2, adjustedY);
        context.rect(adjustedX - fullSize / 2, adjustedY - fullSize / 2, fullSize, fullSize);
    }

    function drawPiece(location, square, hole, white, tall) {
        context.lineWidth = fullRadius / 95.5;
        if (tall) {
            context.lineWidth *= 2.5;
        }
        context.strokeStyle = white ? '#00BB00' : '#002200';
        if (square) {
            drawSquare(location[0], location[1], 0.22);
        } else {
            drawCircle(location[0], location[1], 0.11);
        }
        if (hole) {
            drawCircle(location[0], location[1], 0.08);
        }
    }

})(jQuery);