<script>
  import { Router } from "svelte-router-spa";
  import Home from "./Home.svelte";
  import Room from "./Room.svelte";
  import NavBar from "./NavBar.svelte";
  import Alerts from "./Alerts.svelte";

  const routes = [
    {
      name: "/",
      component: Home
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

  footer {
    width: 100%;
    /* Set the fixed height of the footer here */
  }

  ul {
    text-align: right;
  }
</style>

<NavBar />
<main class="bg-light">
  <Router {routes} />
</main>
<footer>
  <ul class="m-3">
    <li class="mx-3 d-inline">
      <a href="http://github.com/m0ppers/wuerfler">github</a>
    </li>
  </ul>
</footer>

<Alerts />
