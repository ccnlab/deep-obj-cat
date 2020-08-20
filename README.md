# deep-obj-cat

This repository contains simulations and data associated with this paper:

* O’Reilly, R. C., Russin, J. L., Zolfaghar, M., & Rohrlich, J. (2020). Deep Predictive Learning in Neocortex and Pulvinar. ArXiv:2006.14800 [q-Bio]. http://arxiv.org/abs/2006.14800

And (in the future), followup work of relevance to the central themes of this work.

It is organized as follows:

* `expts`: information associated with the experimental tests of the theory, initially a shape comparison study on m-turk where participants compared shapes processed according to the visual front-end used in our model.  This study found that people categorize the objects according to the same shape categories discovered by our self-organizing predictive learning model.

* `papers`: LaTeX and other source versions of the above paper, and prior attempts to publish these ideas.

* `results`: figures, data, analysis scripts for the simulations and experimental data.  Mostly RSA (representational similarity analysis) but also some categorization.  `res.go` contains primary analysis code.

* `sims`: simulations

    + `cemer`: original C++-based emergent (https://github.com/emer/cemer) simulations of the what-where-integration (wwi) model featured in the above paper.  It runs on the latest released version of cemer.
    
    + `prednet`: the prednet comparision models, implemented in pytorch by Jake Russin.
    
    + `wwwi3d`: latest version of the deep predictive learning model, implemented in our new Go-based emergent framework.  This version is easier to understand due to the direct code-based simulation approach, and a version that can be run in Python will be available soon.  However, it is based on an updated version of the deep predictive learning framework that is organized around the closed thalamocortical loops instead of the driver projections, so it differs from the original cemer version from the paper.  It is currently (as of Aug 2020) a work in progress, but does run.
    
    + `obj3denv`: contains Go-based code that generates the 3D object images and saccade eye movements used to train the model.  The original cemer model generates its inputs on the fly, whereas now we are pre-rendering.  A .tar file of the images will be available soon.
    
