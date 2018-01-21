<hardware-list>
  <div class="columns is-centered">
    <div class="column is-4" each={ cpu, index in miner.config.cpus }>
      <div class="card">
        <div class="card-header">
          <p class="card-header-title">
            { cpu.model }
          </p>
          <!--
          <div class="card-header-icon">
            <div class="field">
              <input id="switchRoundedDefault" type="checkbox" name="switchRoundedDefault" class="switch is-rounded" checked="checked">
              <label for="switchRoundedDefault"></label>
            </div>
          </div>
          -->
        </div>
        <div class="card-content has-text-centered">
          <h3 class="title is-3">{ cpuHashrate(index).toFixed(2) }</h3>
          <h3 class="subtitle">H/s</h3>
        </div>
        <div class="card-footer">
          <div class="card-footer-item">
            <nav class="level" style="flex: 1">
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Threads</p>
                  <p class="title">
                    <select class="select" data-index={index} onchange={updateThreads}>
                      <option each={i in threadNums[index]} value={i} selected={cpu.threads == i}>{i}</option>
                    </select>
                  </p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Coin</p>
                  <p class="title">
                    <select class="select" data-index={index} onchange={updateCoin}>
                      <option each={ coin, name in miner.config.coins } value={name} selected={cpu.coin == name}>{ name }</option>
                    </select>
                  </p>
                </div>
              </div>
            </nav>
          </div>
        </div>
      </div>
    </div>

<!--
    <div class="column is-4" each={ miner.hardware.gpus }>
      <div class="card">
        <div class="card-header">
          <p class="card-header-title">
            { model }
          </p>
          <div class="card-header-icon">
            <div class="field">
              <input id="asdasdasd" type="checkbox" class="switch is-rounded" checked="checked">
              <label for="asdasdasd"></label>
            </div>
          </div>
        </div>
        <div class="card-content has-text-centered">
          <h3 class="title is-3">1234</h3>
          <h3 class="subtitle">H/s</h3>
        </div>
        <div class="card-footer">
          <div class="card-footer-item">
            <nav class="level" style="flex: 1">
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Intensity</p>
                  <p class="title">
                    <select class="select">
                      <option>1</option>
                      <option>2</option>
                      <option>3</option>
                      <option>4</option>
                    </select>
                  </p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Coin</p>
                  <p class="title">
                    <select class="select">
                      <option>xmr</option>
                      <option>eth</option>
                    </select>

                  </p>
                </div>
              </div>
            </nav>
          </div>
        </div>
      </div>
    </div>
  -->

  </div>

  <script>
    this.miner = opts.miner

    this.threadNums = []
    this.miner.hardware.cpus.forEach(function (cpu) {
      var threads = []
      for(var thread = 0; thread < cpu.virtual_cores; thread++) {
        threads[thread] = thread+1
      }
      this.threadNums[cpu.index] = threads
    }.bind(this))


    this.stats = []
    this.miner.on('stats', function(stats) {
      this.stats = stats
      this.update()
    }.bind(this))

    cpuHashrate(index) {
      if (this.stats.length == 0) {
        return 0
      }

      var cpuStat = this.stats.cpu_stats.find(function (s) { return s.index == index })

      if(!cpuStat) {
        return 0
      }

      return cpuStat.hashrate
    }

    updateThreads(e) {
      var index = parseInt(e.srcElement.dataset.index)
      var threads = parseInt(e.srcElement.value)
      opts.miner.config.cpus[index].threads = threads
      opts.miner.trigger('update')
    }

    updateCoin(e) {
      var index = parseInt(e.srcElement.dataset.index)
      opts.miner.config.cpus[index].coin = e.srcElement.value
      opts.miner.trigger('update')
    }

  </script>

</hardware-list>
