# shape similarity experiment

* Followed this tutorial: https://blog.mturk.com/tutorial-how-to-label-thousands-of-images-using-the-crowd-bea164ccbefc

* Overall strategy is just to pre-compile all the images, and then present users with a binary choice (Left or Right).

* Mturk sandbox: https://requestersandbox.mturk.com/create/projects
* Mturk login: https://requester.mturk.com/create/projects

* AWS S3 account used to host the images: https://console.aws.amazon.com/billing/home?#/account

    + Instructions for making things public has changed since the tutorial -- I ended up just disabling all the various public ACL kinds of options under Permissions, and then it allowed me to make them public.
    
* `shape-cmp.go` generates the experimental trials, compositing images previously generated using the `imgproc.go` code, which processes image files captured from model, avail for download here: https://grey.colorado.edu/downloads/wwi_emer_imgs_20fg_8tick_rot1.tar.gz
    + Current version: runs image through V1 filters and then inverts back out using random selection of max pool unit -- fairly unrecognizable but definitely gives an overall impression of shape..
    + Previous: turns all the blue background color to white, and everything else to black, and then blurs the resulting image with a bild blur.Box convolution filter with a radius of 25.  Produces pretty blurry images.


