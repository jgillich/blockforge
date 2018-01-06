<app>
  <div>
    <nav class="navbar is-primary">
      <div class="navbar-brand">
        <div class="navbar-start">
          <div class="navbar-item">
            <img src="https://bulma.io/images/bulma-logo.png" width="112" height="28">
          </div>
        </div>
      </div>
      <div class="navbar-end">
        <div class="navbar-item">
          <div class="field is-grouped">
            <p class="control">
              <a class="button is-success">
                <span class="icon">
                  <i class="fa fa-play"></i>
                </span>
                &nbsp;Start
              </a>
            </p>
          </div>
        </div>
      </div>
    </nav>
  </div>
  <section class="section">
    <div class="container">
      <cpu items={opts.config.cpu}>
    </div>
  </section>
</app>
