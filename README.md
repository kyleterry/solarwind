# Solarwind

Solarwind is the for-reals no-bullshit static site generator.

Static site generators have become confusing and bloated. I've tried 4 or 5 and
they don't make sense, they have terrible defaults, the templates are hard to
understand, yadda yadda nerd rage yadda.

This will be a simple and easy to use SSG.

First things first, you will need a `Solarwindfile`. Put this shit in your site
root:

```
{
    "title": "My Site Title",
    "description: "This is a description that will be in my meta tags"
}
```

It's fucking JSON. How about that?

Next you will need Solarwind:

`go get github.com/kyleterry/solarwind`

This shit will show up in your `$GOPATH/bin`.

Now you will need to copy the starter template into your site root:

`cp -r $GOPATH/src/github.com/kyleterry/solarwind/starter/templates /my/fucking/site/path/`

Now create a content and posts dir:

`mkdir -p /my/fucking/site/path/content/posts`

YAY

Now generate your site:

`solarwind generate`

and all your shit shows up in `/my/fucking/site/path/public`

If you need static assets, just put them in `/my/fucking/site/path/static/{css,js,images}`
or whatever (really, I just copy that entire dir to `/my/fucking/site/path/public/static`).

K... Have fun.
