<script>
  import { Nav, NavItem, Navbar, NavLink, NavbarBrand } from "sveltestrap";
</script>

<header>
  <Navbar color="dark" dark>
    <NavbarBrand href="/">WÜRFLER</NavbarBrand>
    <Nav class="ml-auto" navbar>
      <NavItem>
        <NavLink class="px-2" href="/datenschutz">Datenschutzerklärung</NavLink>
      </NavItem>
      <NavItem>
        <NavLink class="px-2" href="https://github.com/m0ppers/wuerfler">
          GitHub
        </NavLink>
      </NavItem>
    </Nav>
  </Navbar>
</header>
