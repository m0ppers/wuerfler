<script>
  import { Alert } from "sveltestrap";
  import { alerts } from "./stores";
</script>

<style>
  div {
    position: absolute;
    top: 10px;
    right: 10px;
    width: 25vw;
  }
</style>

<div>
  {#each $alerts as alert}
    <Alert color={alert.color}>{alert.text}</Alert>
  {/each}
</div>
