<script lang="typescript">
    import geil from "./images/geil.png";

    const src = geil;
</script>

<div class="outline">
    <img {src} class="circle" />
</div>
<img {src} class="circle" />

<style type="scss">
    .outline {
        border: 1px grey dashed;
    }
    .circle {
        width: 100%;
        height: 100%;
        background: red;
        border-radius: 50%;
    }
</style>
