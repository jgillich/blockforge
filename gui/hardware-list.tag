<hardware-list>
  <div class="columns is-centered">

    <div class="column is-4" each={ miner.config.processors }>
      <div class="card">
        <div class="card-header">
          <p class="card-header-title" style="white-space: nowrap; text-overflow: ellipsis; overflow: hidden">
            { name }
          </p>
        </div>
        <div class="card-content has-text-centered">
          <h3 class="title is-3">{ hashrate(index).toFixed(2) }</h3>
          <h3 class="subtitle">H/s</h3>
        </div>
        <div class="card-footer">
          <div class="card-footer-item">
            <nav class="level" style="flex: 1">
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Enabled</p>
                  <p class="title">
                    <input id={ "enable" + index} type="checkbox" class="switch is-rounded" data-index={index} checked={ enable } onclick={ toggleEnable }>
                    <label for={ "enable" + index}></label>
                  </p>
                </div>
              </div>

              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Threads</p>
                  <p class="title">
                    <select class="select" data-index={index} onchange={updateThreads}>
                      <option each={i in threadNums[index]} value={i} selected={threads == i}>{i}</option>
                    </select>
                  </p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Coin</p>
                  <p class="title">
                    <select class="select" data-index={index} onchange={updateCoin}>
                      <option each={ c, name in miner.config.coins } value={name} selected={coin == name}>{ name }</option>
                    </select>
                  </p>
                </div>
              </div>
            </nav>
          </div>
        </div>
      </div>
    </div>

    <div class="column is-4" each={ miner.config.opencl_devices }>
      <div class="card">
        <div class="card-header">
          <p class="card-header-title">
            { name }
          </p>
        </div>
        <div class="card-content has-text-centered">
          <h3 class="title is-3">{ hashrate(index, platform).toFixed(2) }</h3>
          <h3 class="subtitle">H/s</h3>
        </div>
        <div class="card-footer">
          <div class="card-footer-item">
            <nav class="level" style="flex: 1">
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Enabled</p>
                  <p class="title">
                    <input id={ "enable" + platform + index } type="checkbox" class="switch is-rounded"
                       data-index={ index } data-platform={ platform } checked={ enable } onclick={ toggleEnable }>
                    <label for={ "enable" + platform + index }></label>
                  </p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Intensity</p>
                  <p class="title">
                    <select class="select" data-index={index} data-platform={platform} onchange={updateIntensity}>
                      <option each={ i in this.intensities } value={i} selected={intensity == i}>{ i }</option>
                    </select>
                  </p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Coin</p>
                  <p class="title">
                    <select class="select" data-index={index} data-platform={platform} onchange={updateCoin}>
                      <option each={ c, name in miner.config.coins } value={name} selected={coin == name}>{ name }</option>
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
    this.intensities = []
    for (i = 64; i < 1000; i += 64) {
      this.intensities.push(i)
    }

    this.threadNums = []
    opts.miner.processors.forEach(function (cpu) {
      var threads = []
      for(var thread = 0; thread < cpu.virtual_cores; thread++) {
        threads[thread] = thread+1
      }
      this.threadNums[cpu.index] = threads
    }.bind(this))

    opts.miner.on('stats', function(stats) {
      this.stats = stats
      this.update()
    }.bind(this))

    processor(i) {
      return opts.miner.config.processors.find(function(p) {
        return p.index == i
      })
    }

    cl(i, p) {
      return opts.miner.config.opencl_devices.find(function(d) {
        return d.index == i && d.platform == p
      })
    }

    hashrate(index, platform) {
      if (!this.stats) {
        return 0
      }

      if (platform == undefined) {
        var processor = this.processor(index)
        if(!processor.enable) return 0
        var hps = 0
        for(var i = 0; i < processor.threads; i++) {
          hps += this.stats["worker.cpu." + index + "." + i] || 0
        }
        return hps
      } else {
        var cl = this.cl(index, platform)
        if(!cl.enable) return 0
        return this.stats["worker.opencl." + index + "." + platform] || 0
      }
    }

    updateThreads(e) {
      var index = parseInt(e.srcElement.dataset.index)
      var threads = parseInt(e.srcElement.value)
      var processor = this.processor(index)

      processor.threads = threads
      opts.miner.trigger('update')
    }

    updateCoin(e) {
      var index = parseInt(e.srcElement.dataset.index)
      var platform = parseInt(e.srcElement.dataset.platform)

      if (isNaN(platform)) {
        var processor = this.processor(index)
        processor.coin = e.srcElement.value
      } else {
        var cl = this.cl(index, platform)
        cl.coin = e.srcElement.value
      }

      opts.miner.trigger('update')
    }

    updateIntensity(e) {
      var index = parseInt(e.srcElement.dataset.index)
      var platform = parseInt(e.srcElement.dataset.platform)

      if (isNaN(e.srcElement.value)) {
        alert("intensity must be a number")
        return
      }

      var intensity = parseInt(e.srcElement.value, 10)

      if (intensity > 2000) {
        alert("intensity cannot be larger than 2000")
        return
      }

      var cl = this.cl(index, platform)
      cl.intensity = intensity
      opts.miner.trigger('update')
    }

    toggleEnable(e) {
      var index = parseInt(e.srcElement.dataset.index)
      var platform = parseInt(e.srcElement.dataset.platform)

      if (isNaN(platform)) {
        var processor = this.processor(index)
        processor.enable = !processor.enable
      } else {
        var cl = this.cl(index, platform)
        cl.enable = !cl.enable
      }

      opts.miner.trigger('update')
    }

  </script>

</hardware-list>
