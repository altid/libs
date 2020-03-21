# [altid.github.io](https://altid.github.io/)

# Complexity Tradeoffs

> Controlling complexity is the essence of computer programming \- Brian Kernighan

As stated in the basic setup guide, Altid uses a three\-tiered system, connecting clients to services through servers.

This is a powerful abstraction, and is used by many solutions. The client only has to know how to talk to the server, and the services only have to know how to talk to the server.

## State Of The Art

In a modern computing environment, services are expressed in a multitude of ways. Each particular service generally has a native client for interaction, and\/or a web\-based SaaS solution. Each comes with its own nuanced interface and control scheme, though many efforts have been made at homogenous interfaces for multiple services. For example, Emacs has handily solved this problem, for many users; providing a single, homogenous interface to multiple services.

This approach pushes much of the complexity to the client implementation. Either it has to be very generalized, and able to interface well with nearly arbitrary service implementations, as is the case with Emacs and a web browser; or it’s a singular purpose, bespoke client which may take many thousands of hours to perfect, and make bug free. Additionally, as with many things related to computing, the very dynamic landscape leads to new bugs that must be dealt with, for each client.

## Altid’s Complexity Model

One of the initial goals of Altid was to defer a majority of computation to a centralized server in one’s home.

All of the state for a given session is a marriage of the services, and the server. The services provide a series of directories, and the server aggregates content from the directories and presents it to the client \(this is described in more detail on the [architectural overview](/architecture.html) page\). A bug introduced that would affect any running service would require work only on that specific service \- no modifications to the server or any client would be needed.

## Single Buffer View

An Altid client is served a single view, representing a single buffer of a single service \(For more information about buffers and services, refer to the [overview](/). It can request to switch to other buffers, and also has a list of all available buffers via tabs; but fundamentally a single connection to a server results in a single buffer view.

This may seem limiting, but due to the very simple, stable client implementations that result, clients may wish to issue multiple connections, to multiple services at once, providing for a very granular approach to setting up a client. For example, one could connect to Discord, IRC, sms, Slack, email, or any number of other chats in a single client window. Depending on the client implementation, you would see an aggregate of all tabs opened.

## Homogenous Interface

One of the benefits of this method of aggregating services to a client is, much like what is achieved in Emacs you realise a single interace to interact with any number of clients. The available commands will depend on the underlying service, and the particular UI elements present may change, but the way you interact with everything is homogenized.

## Aside \- A Central Theme

Due to each service only providing a markdown representation of its state, the eventual theme is client\-dependent. This is nice from an aesthetic perspective, but it’s very powerful when met with any of the following:

 - non\-standard input devices, such as braille readers
 - providing content to vision impaired persons \(high contrast themes, large font\)
 - screen reading solutions

## Complexity

Many of the things mentioned above would not be simple with a different model for complexity, though many things are very difficult this way.

 - Access to service state is very trivial, at the expense of the content being provided in a potentially very lossy manner; a well\-formatted document loses the individuality of font selection, precision page layout, and many facets of branding
 - Aggregating services to a single client is very trivial, at the expense of multiple connections
 - Service implementations are trivial to write, at the expense of requiring the programmer to define a state as a set of directories of particular files, which is potentially very limiting
 - Service and client implementations are trivial to write, at the expense of more complex servers

## Author’s Note

All of these tradeoffs are less than ideal, but in practice the lossy nature allows me to read a document on my phone, walk over to my computer, and continue reading instantly; the same for any Altid service. Defining a state, \(not as a relationship with a client\), as a standalone entity for sessions is what many services on the web are accomplishing, but the bar to entry to provide such a thing is high, expensive, and resource\-intensive.

If you wish to discuss this, or anything else, feel free to join us at \#altid on Freenode. I’m always happy to hear opinions, suggestions, and feedback\!

