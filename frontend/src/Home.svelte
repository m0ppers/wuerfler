<script>
  import {
    Form,
    FormGroup,
    FormText,
    Input,
    Label,
    Button,
    Jumbotron
  } from "sveltestrap";
  import axios from "axios";

  let roomName = "";

  const handleSubmit = async e => {
    e.preventDefault();
    // new String so axios thinks this is an object and does JSON encoding instead of form post
    // no idea how to do that nonhackish
    const createdName = await axios.post("/api/rooms", new String(roomName));
    location.assign(
      `//${location.host}/rooms/${encodeURIComponent(createdName.data)}`
    );
  };
</script>

<style>
  .flexi {
    display: inline-flex;
  }
</style>

<Jumbotron class="text-center bg-light">
  <div class="container">
    <h1>WÜRFLER</h1>
    <p class="lead text-muted">
      Your online RPG Companion to roll dices with your party
    </p>
    <Form on:submit={handleSubmit}>
      <FormGroup>
        <div class="creategroup">
          <Label>Create a room</Label>
          <div>
            <div class="flexi">
              <Input
                placeholder="Room name"
                readonly={false}
                bind:value={roomName} />
            </div>
            <div class="flexi">
              <Button class="primary" type="submit">Create</Button>
            </div>
          </div>
        </div>
      </FormGroup>
    </Form>
  </div>
</Jumbotron>
