# az-wrap

Just some pretty printing and upcoming TUI work around the Azure CLI. Mostly developing it for fun... And personal usefulness

## What job, exactly?

I work with a lot of Azure subscriptions, and in one tenant, sequence numbers or GUIDs are used instead of meaningful names.

This means, I'd have to keep track of the subscription sequence numbers in my head. Instead, using this dumb tool, I can create aliases for the subscriptions I have access to.

## Example output

![example_image](assets/example_output.png)

## Usage

Get the menu by running `az-wrap`, or create an alias for it in your
bashrc, zshrc, fish, powershell profile. Since it only has one function right now, my alias is simply set to `subs`.

### Set an alias

Set an alias by passing `-alias` flag.

`az-wrap -alias subscriptionId:alias`
