<cpu>
  <div class="columns">
    <div class="column is-4" each={ cpu, name in opts.items }>
      <div class="card">
        <div class="card-header">
          <p class="card-header-title">
            { name }
          </p>
        </div>
        <div class="card-content has-text-centered">
          <div class="title is-3">1234 H/s</div>
        </div>
        <div class="card-footer">
          <span class="card-footer-item">
            <div class="field is-horizontal">
              <div class="field-label is-medium">
                <label class="label">Threads</label>
              </div>
              <div class="field-body">
                <div class="field">
                  <div class="control">
                    <select class="select">
                      <option>1</option>
                      <option>2</option>
                      <option>3</option>
                      <option>4</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>
          </span>
          <span class="card-footer-item">
              <div class="field is-horizontal">
                <div class="field-label is-medium">
                  <label class="label">Coin</label>
                </div>
                <div class="field-body">
                  <div class="field">
                    <div class="control">
                      <select class="select">
                        <option>xmr</option>
                        <option>eth</option>
                        <option>3</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>
            </span>
        </div>
      </div>
    </div>
  </div>

</cpu>

