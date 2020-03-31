<script>
  import { Button, Card, CardBody, CardTitle } from "sveltestrap";
  import PlayerSettings from "./PlayerSettings.svelte";
  import { createEventDispatcher } from "svelte";
  import { friends } from "./stores.js";

  const dispatch = createEventDispatcher();

  let hand = [];

  const dices = [4, 6, 8, 10, 12, 20, 100];

  const addDice = dice => {
    hand = [...hand, dice];
  };

  const removeDice = removeIndex => {
    hand = hand.filter((_, index) => index !== removeIndex);
  };

  const roll = () => dispatch("roll", hand);
</script>

<h2>Player Info</h2>
<Card class="box-shadow">
  <CardBody>
    <PlayerSettings on:profileUpdate />
    <h6 class="text-muted">
      Friends connected: {$friends.length > 0 ? $friends.join(', ') : '-'}
    </h6>
    <h5>Select some dices</h5>
    {#each dices as dice}
      <Button class="m-1" on:click={e => addDice(dice)}>d{dice}</Button>
    {/each}
    <h5 class="mt-3">Selected dices</h5>
    <div class="font-italic">Click dice to remove</div>

    <div>
      <!-- hacky...don't let the layout jump when the player selected some dices -->
      {#if hand.length == 0}
        <Button
          class="m-1 invisible"
          color="secondary"
          on:click={e => removeDice(i)}>
          d1
        </Button>
      {:else}
        {#each hand as dice, i}
          <Button class="m-1" color="secondary" on:click={e => removeDice(i)}>
            d{dice}
          </Button>
        {/each}
      {/if}
      <Button
        disabled={hand.length == 0}
        color="primary"
        on:click={roll}
        class="d-block mt-3">
        ROLL IT
      </Button>
    </div>
  </CardBody>

</Card>
