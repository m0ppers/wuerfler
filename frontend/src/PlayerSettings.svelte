<script>
  import { createEventDispatcher } from "svelte";
  import { Form, Input, Button } from "sveltestrap";
  import pencil from "bootstrap-icons/icons/pencil.svg";

  import { myself, friends } from "./stores.js";

  const dispatch = createEventDispatcher();
  let nameInput = "";
  let editing = false;

  const editName = () => {
    nameInput = $myself;
    editing = true;
  };

  const handleEdit = e => {
    e.preventDefault();
    editing = false;
    myself.update(() => nameInput);

    dispatch("profileUpdate", nameInput);
  };
</script>

<style type="scss">
  .outline {
    width: 50%;
    border: 1px grey dashed;
  }
  .circle {
    width: 100%;
    height: 100%;
    background: red;
    border-radius: 50%;
  }

  .icon {
    display: inline;
    cursor: pointer;
  }

  .name {
    font-size: 24px;
  }

  .gametable {
    width: 100%;
  }
</style>

<div class="gametable">
  <div>
    <p class="m-0 font-italic">Your name</p>
    {#if editing}
      <Form on:submit={handleEdit}>
        <Input bind:value={nameInput} />
        <Button class="primary" type="submit">Edit</Button>
      </Form>
    {:else}
      <p class="name">
        {$myself}
        <img
          class="icon"
          src={pencil}
          alt="edit"
          width="24"
          height="24"
          on:click={editName} />
      </p>
    {/if}
  </div>
</div>
