if (typeof console === 'undefined') {
  console = {error:function(){},log:function(){}};
}

var s; // our socket
var messageBody = document.querySelector('#messages_body ul.messages');
var composer = document.querySelector('#messages_body .composer');
var composerInputField = composer.querySelector('.body');
var composingMessages = {}; // room -> string
var accessKeys = {};
var unseenMessageCount = {}; // room -> int


function makeMessageView(message) {
  var li = document.createElement('li');
  var author = document.createElement('span');
  author.innerText = message.author;
  author.className = 'author';
  li.appendChild(author);
  var body = document.createElement('span');
  body.innerText = message.body;
  body.className = 'body';
  li.appendChild(body);
  return li;
}

function showListItems(list, makefun, replaceExisting, items) {
  list.style.display = 'none';
  if (replaceExisting) { list.innerText = ''; }
  if (Array.isArray(items)) {
    items.forEach(function (item) {
      list.appendChild(makefun(item));
    });
  } else if (items) {
    Object.keys(items).forEach(function (k) {
      list.appendChild(makefun(items[k]));
    });
  }
  list.style.display = null; 
}

function showMessages(replaceExisting, messages) {
  showListItems(messageBody, makeMessageView, replaceExisting, messages);
}




gotalk.handleNotification('newmsg', function (m) {
  console.log(m)
  showMessages(false, [m]);
});



gotalk.handleNotification('showmessages', function (messages) {

  makeMessageView("Deneme");
  if (messages) {
    Object.keys(messages).forEach(function (k) {
      showMessages(/*replaceExisting=*/false, [messages[k]]);
  
    });
    console.info("mesajlar geldi")
    console.log(messages)

  }
});

// We get assigned a username
gotalk.handleNotification('username', function (username) {
  Array.prototype.forEach.call(document.querySelectorAll('.my-username'), function (e) {
    e.innerText = username;

    console.log("Hoşgeldin:"+username)
  });
});


// mesajı alıyor
function getComposerMessage() {
  return composerInputField.value.replace(/^[ \s\t\r\n]+|[ \s\t\r\n]+$/g, '');
}


composer.onsubmit = function () {
  console.log("submit eventi çalıştı")
  var body = getComposerMessage();
  console.log("Mesaj:"+body)

  s.request('send-message', { message:{body:body}}, function (err, res) {
    composerInputField.value = "";

  });
  return false;
};

function onConnect(s) {
  console.log("connection opened")
}

var s = gotalk.connection()
  .on('open', onConnect)
  .on("close", err => {
    console.log("connection closed" + (err ? " with error: " + err : ""))
  })


document.addEventListener('keydown', function (ev) {
  if (ev.ctrlKey) {
    var f = accessKeys[String.fromCharCode(ev.keyCode).toUpperCase()];
    if (f) {
      ev.preventDefault();
      ev.stopPropagation();
      f(ev);
    }
  }
}, true);
