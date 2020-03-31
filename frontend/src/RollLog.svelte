<script lang="typescript">
  import {
    Button,
    Card,
    CardBody,
    CardFooter,
    CardHeader,
    CardImg,
    CardSubtitle,
    CardText,
    CardTitle
  } from "sveltestrap";

  import { rolls } from "./stores.js";
</script>

<h2>Roll log</h2>
{#each $rolls as roll}
  <Card class="mb-3">
    <CardBody>
      <h6 class="mb-0">
        {roll.name}
        <small class="text-muted">{new Date(roll.date).toLocaleString()}</small>
      </h6>
      <p>
        {#each roll.results as rollResult}
          <span class="badge badge-pill badge-primary mx-1">
            d{rollResult.dice}: {rollResult.result}
          </span>
        {/each}
      </p>
    </CardBody>
  </Card>
{/each}
