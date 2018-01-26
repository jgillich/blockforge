<hardware-list>
  <div class="columns is-centered">

    <div class="column is-4" each={ processor, i in miner.config.processors }>
      <div class="card">
        <div class="card-header">
          <p class="card-header-title">
            { processor.name }
          </p>
          <div class="card-header-icon">
            <div class="field">
              <input id={ "pswitch" + i} type="checkbox" class="switch is-rounded" checked={ processor.enable } onclick={ toggleEnable }>
              <label for={ "pswitch" + i}></label>
            </div>
          </div>
        </div>
        <div class="card-content has-text-centered">
          <h3 class="title is-3">{ cpuHashrate(i).toFixed(2) }</h3>
          <h3 class="subtitle">H/s</h3>
        </div>
        <div class="card-footer">
          <div class="card-footer-item">
            <nav class="level" style="flex: 1">
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Threads</p>
                  <p class="title">
                    <select class="select" data-index={i} onchange={updateThreads}>
                      <option each={i in threadNums[processor.index]} value={i} selected={threads == i}>{i}</option>
                    </select>
                  </p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Coin</p>
                  <p class="title">
                    <select class="select" data-index={index} onchange={updateCoin}>
                      <option each={ c, name in miner.config.coins } value={name} selected={processor.coin == name}>{ name }</option>
                    </select>
                  </p>
                </div>
              </div>
            </nav>
          </div>
        </div>
      </div>
    </div>

    <div class="column is-4" each={ cl, i in miner.config.opencl_devices }>
      <div class="card">
        <div class="card-header">
          <p class="card-header-title">
            { cl.name }
          </p>
          <div class="card-header-icon">
            <div class="field">
              <input id={ "clswitch" + i} type="checkbox" class="switch is-rounded" checked={ cl.enable } onclick={ toggleEnable }>
              <label for={ "clswitch" + i}></label>
            </div>
          </div>
        </div>
        <div class="card-content has-text-centered">
          <h3 class="title is-3">0.00</h3>
          <h3 class="subtitle">H/s</h3>
        </div>
        <div class="card-footer">
          <div class="card-footer-item">
            <nav class="level" style="flex: 1">
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Intensity</p>
                  <p class="title">
                    <select class="select" data-index={i} onchange={updateIntensity}>
                      <option value="1">1</option>
                    </select>
                  </p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Coin</p>
                  <p class="title">
                    <select class="select" data-index={index} onchange={updateCoin}>
                      <option each={ c, name in miner.config.coins } value={name} selected={cl.coin == name}>{ name }</option>
                    </select>
                  </p>
                </div>
              </div>
            </nav>
          </div>
        </div>
      </div>
    </div>

  </div>

  <script>
    this.miner = opts.miner

    this.threadNums = []
    this.miner.processors.forEach(function (cpu) {
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
      opts.miner.config.processors[index].threads = threads
      opts.miner.trigger('update')
    }

    updateCoin(e) {
      var index = parseInt(e.srcElement.dataset.index)
      opts.miner.config.processors[index].coin = e.srcElement.value
      opts.miner.trigger('update')
    }

    updateIntensity(e) {

    }

    toggleEnable(e) {
      var id = e.srcElement.id

      if (id.includes("pswitch")) {
        var pid = parseInt(id.replace("pswitch", ""), 10)
        opts.miner.config.processors[pid].enable = !opts.miner.config.processors[pid].enable
        opts.miner.trigger('update')
      } else if (id.includes("clswitch")) {
        var clid = parseInt(id.replace("clswitch", ""), 10)
        opts.miner.config.opencl_devices[clid].enable = !opts.miner.config.opencl_devices[clid].enable
        opts.miner.trigger('update')
      }

      this.update()
    }

  </script>

</hardware-list>
