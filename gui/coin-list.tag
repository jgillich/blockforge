<coin-list>
  <div class="card">
    <div class="card-header">
      <p class="card-header-title">
        Coins
      </p>
    </div>
    <div class="card-table">
      <table class="table is-fullwidth is-striped">
        <tbody>
          <tr class="is-size-5" each={ coin, name in miner.config.coins }>
            <td>
              <i class={ "cc " + name}></i>
            </td>
            <td>{ name }</td>
            <td>{ coin.pool.url }</td>
            <td>{ coin.pool.user.substring(0, 20) }</td>
            <td>
              <a class="button is-small is-primary" data-coin={name} onclick={edit}>Edit</a>
              <a class="button is-small is-danger" data-coin={name} onclick={deleteCoin}>Delete</a>
            </td>
          </tr>

        </tbody>
      </table>
    </div>

    <footer class="card-footer" show={true}> <!-- TODO hide when all coins are added -->
      <a href="#" class="card-footer-item" onclick={ add }>Add Coin</a>
    </footer>
  </div>

  <div if={ showEdit }>
    <coin-edit close={ closeEdit } miner={miner} coin={editItem}></coin-edit>
  </div>

  <script>
    this.miner = opts.miner
    this.showEdit = false
    this.editItem = false

    edit (e) {
      this.editItem = e.target.dataset.coin
      this.showEdit = true
      this.update()
    }

    add() {
      this.editItem = false
      this.showEdit = true
      this.update()
    }

    closeEdit() {
      this.showEdit = false
      this.update()
    }

    // TODO go through hardware and reset coin
    deleteCoin(e) {
      delete opts.miner.config.coins[e.target.dataset.coin]
      opts.miner.trigger('update')
    }
  </script>
</coin-list>
