<script>
  import Sidebar from "./Sidebar.svelte";
  import RollLog from "./RollLog.svelte";
  import { alerts, myself, friends, rolls } from "./stores.js";
  import { onDestroy } from "svelte";

  export let currentRoute;

  let url = `${location.host}/rooms/${currentRoute.namedParams.name}/websocket`;
  let protocol = location.protocol == "https:" ? "wss" : "ws";

  const ws = new WebSocket(`${protocol}://${url}`);

  const unsubscribe = myself.subscribe(value => {
    if (value && value !== "") {
      localStorage.setItem("name", value);
    }
  });

  onDestroy(unsubscribe);

  ws.onopen = () => {
    ws.send(
      JSON.stringify({
        type: "join",
        payload: localStorage.getItem("name") || ""
      })
    );
  };

  ws.onmessage = message => {
    message = JSON.parse(message.data);
    switch (message.type) {
      case "usersupdate":
        myself.set(message.payload.self);
        friends.set(message.payload.others);
        break;
      case "roll":
        rolls.update(rolls => [message.payload, ...rolls.slice(0, 49)]);
        break;
    }
  };

  ws.onclose = () => {
    alerts.update(oldAlerts => [
      ...oldAlerts,
      { text: "Lost connection to websocket", color: "danger" }
    ]);
    console.log("CLOSING", arguments);
  };

  ws.onerror = err => {
    console.error(err);
    alerts.update(oldAlerts => [...oldAlerts, err]);
  };

  const roll = event => {
    ws.send(JSON.stringify({ type: "roll", payload: event.detail }));
  };

  const profileUpdate = event => {
    ws.send(JSON.stringify({ type: "profileUpdate", payload: event.detail }));
  };
</script>

<div class="container">
  <div class="row">
    <h2 class="display-4">
      {decodeURIComponent(currentRoute.namedParams.name)}
    </h2>
  </div>
  <div class="row">
    <div class="col-sm">
      <Sidebar on:roll={roll} on:profileUpdate={profileUpdate} />
    </div>
    <div class="col-sm">
      <RollLog />
    </div>
  </div>
</div>
