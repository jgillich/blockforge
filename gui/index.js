var h = picodom.h
var patch = picodom.patch

var node

if (!window.miner) {
  window.miner = {}
  miner.data = { "config": { "donate": 1, "log_level": "", "coin": { "xmr": { "pool": { "url": "stratum+tcp://xmr.poolmining.org:3032", "user": "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr", "pass": "x" } } }, "cpu": { "Intel Core i5-6200U": { "coin": "xmr", "threads": 1 } }, "gpu": null } }

}

render(view, miner.data)

function render(view, state) {
  patch(node, (node = view(state)))
}


function view(state) {
  return (
    h('div', {},
      h('nav', { class: 'navbar is-primary' },
        h('div', { class: 'navbar-start' },
          h('div', { class: 'navbar-brand' },
            h('span', { class: 'navbar-item' },
              h('img', { src: 'https://bulma.io/images/bulma-logo.png', width: "112", height: "28" })
            )
          )
        ),
        h('div', { class: 'navbar-end' },
          h('div', { class: 'navbar-item' },
            h('div', { class: 'field is-grouped' },
              h('p', { class: 'control' },
                h('a', { class: 'button is-primary' },
                  h('span', { class: 'icon' },
                    h('i', { class: 'fa fa-play' })
                  ),
                  h('span', {}, 'Start')
                )
              )
            )
          )
        )
      ),
      h("section", { class: "section" },
        h("div", { class: 'container' },
          cpuView(state.config)
        )
      )
    )
  )
}



function cpuView(config) {
  return h('div', { class: 'columns' },
    Object.keys(config.cpu).map(function (key) {
      return (
        h('div', { class: 'column is-4' },
          h("div", { class: 'card' },
            h('div', { class: 'card-header' },
              h('p', { class: 'card-header-title' }, key)
            ),
            h('div', { class: 'card-content has-text-centered' },
              h('h3', { class: 'title is-3' }, '1234 H/s')
            ),
            h('footer', { class: 'card-footer' },
              h('span', { class: 'card-footer-item' },
                h('div', { class: 'field is-horizontal' },
                  h('div', { class: 'field-label is-medium' },
                    h('label', { class: 'label' }, 'Threads')
                  ),
                  h('div', { class: 'field-body' },
                    h('div', { class: 'field' },
                      h('div', { class: 'control' },
                        h('div', { class: 'select' },
                          h('select', {},
                            h('option', {}, '1'),
                            h('option', {}, '22'),
                            h('option', {}, '3'),
                            h('option', {}, '4')
                          )
                        )
                      )
                    )
                  )
                )
              ),
              h('span', { class: 'card-footer-item' },
                h('div', { class: 'field is-horizontal' },
                  h('div', { class: 'field-label is-medium' },
                    h('label', { class: 'label' }, 'Coin')
                  ),
                  h('div', { class: 'field-body' },
                    h('div', { class: 'field' },
                      h('div', { class: 'control' },
                        h('div', { class: 'select' },
                          h('select', {},
                            h('option', {}, 'xmr'),
                            h('option', {}, 'eth'),
                          )
                        )
                      )
                    )
                  )
                )
              )
            )
          )
        )
      )
    })
  )
}
