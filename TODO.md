# TODO:

- [ ] Lazily load content if cache miss - Check https://data-star.dev/examples/lazy_load
- [ ] Add E2E tests - Check https://playwright-community.github.io/playwright-go/
- [ ] remove time.Sleep(time.Second \* 2)
- [ ] Fix charts
- [ ] Set cache instead of invalidating them
- [ ] Generate monthly trend on the fly instead of caching it
- [ ] Fix monthly and total click count skeleton
- [ ] Go back to postcss with purgecss: https://chat.deepseek.com/a/chat/s/3b610bc1-eb24-4e64-8f7e-e1109905df9f
- [ ] need to add safelist to purgecss
      bunx purgecss --css public/styles.css --content ../internal/\*_/_.templ --output public/dist -s datastar-swapping -s datastar-settling
- [ ] use tailwindcss bunx instead of devbox package, need to check docker build afterwards
- [ ] remove go charts
- [ ] check about page transition api
- [ ] update readme to install direnv and run following. refer(https://www.jetify.com/docs/devbox/ide_configuration/direnv/)
      devbox generate direnv --env-file .env
- [ ] update readme: refer https://www.jetify.com/docs/devbox/ide_configuration/vscode/ for devbox with vscode or even with other IDEs
