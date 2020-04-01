<script>
  import { Router } from "svelte-router-spa";
  import Home from "./Home.svelte";
  import Datenschutz from "./Datenschutz.svelte";
  import Room from "./Room.svelte";
  import NavBar from "./NavBar.svelte";
  import Alerts from "./Alerts.svelte";

  const routes = [
    {
      name: "/",
      component: Home
    },
    {
      name: "/datenschutz",
      component: Datenschutz
    },
    {
      name: "/rooms/:name",
      component: Room
    }
  ];
</script>

<style>
  main {
    flex: 1;
  }
</style>

<NavBar />
<main class="bg-light">
  <Router {routes} />
</main>

<Alerts />
