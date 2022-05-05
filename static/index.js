let socket = null;
let o = document.getElementById("output");
let userField = document.getElementById("username");
let messageField = document.getElementById("message");

window.onbeforeunload = () => {
  console.log("buh-bye!");
  let jsonData = {};
  jsonData["action"] = "left";
  socket.send(JSON.stringify(jsonData));
};

document.addEventListener("DOMContentLoaded", function () {
  //sockets have 4 relevant methods: open, close, error, message
  socket = new ReconnectingWebSocket("ws://127.0.0.1:8080/ws", null, {
    debug: true,
    reconnectInterval: 3000,
  });

  const offline = `<span class="badge bg-danger">Not Connected</span>`;
  const online = `<span class="badge bg-success">Connected</span>`;
  let statusDiv = document.getElementById("status");

  socket.onopen = () => {
    console.log("Connected to websocket!!");
    statusDiv.innerHTML = online;
  };

  socket.onclose = () => {
    console.log("Closed websocket connection!!");
    statusDiv.innerHTML = offline;
  };

  socket.onerror = (err) => {
    console.log("Websocket Error", err);
    statusDiv.innerHTML = offline;
  };

  socket.onmessage = (msg) => {
    let data = JSON.parse(msg.data);
    console.log("Action is", data.action);

    switch (data.action) {
      case "list_users":
        let ul = document.getElementById("online_users");
        while (ul.firstChild) ul.removeChild(ul.firstChild);

        if (data.connected_users.length > 0) {
          data.connected_users.forEach(function (item) {
            let li = document.createElement("li");
            li.appendChild(document.createTextNode(item));
            ul.appendChild(li);
          });
        }
        break;

      case "broadcast":
        o.innerHTML = o.innerHTML + data.message + "<br>";
        break;
    }
  };

  userField.addEventListener("change", function () {
    let jsonData = {};
    jsonData["action"] = "username";
    jsonData["username"] = this.value;
    socket.send(JSON.stringify(jsonData));
  });

  messageField.addEventListener("keydown", function (event) {
    if (event.code === "Enter") {
      if (!socket) {
        console.log("no connection");
        return false;
      }

      if (userField.value === "" || messageField.value === "") {
        errorMessage("Fill out username and message!");
        return false;
      } else {
        sendMessage();
      }

      event.preventDefault();
      event.stopPropagation();
    }
  });

  document.getElementById("sendBtn").addEventListener("click", function () {
    if (userField.value === "" || messageField.value === "") {
      errorMessage("Fill out username and message!");
      return false;
    } else {
      sendMessage();
    }
  });
});

const sendMessage = () => {
  let jsonData = {};
  jsonData["action"] = "broadcast";
  jsonData["username"] = userField.value;
  jsonData["message"] = messageField.value;
  socket.send(JSON.stringify(jsonData));
  messageField.value = "";
};
const errorMessage = (msg) => {
  notie.alert({
    type: "error",
    text: msg,
  });
};
