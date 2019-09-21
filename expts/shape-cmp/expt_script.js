<!-- You must include this JavaScript file -->
<script src="https://assets.crowd.aws/crowd-html-elements.js"></script>

<!-- For the full list of available Crowd HTML Elements and their input/output documentation,
      please refer to https://docs.aws.amazon.com/sagemaker/latest/dg/sms-ui-template-reference.html -->

<!-- You must include crowd-form so that your task submits answers to MTurk -->
<crowd-form answer-format="flatten-objects">

    <!-- The crowd-classifier element will create a tool for the Worker to select the
           correct answer to your question.

          Your image file URLs will be substituted for the "image_url" variable below 
          when you publish a batch with a CSV input file containing multiple image file URLs.
          To preview the element with an example image, try setting the src attribute to
          "https://s3.amazonaws.com/cv-demo-images/one-bird.jpg" -->
    <crowd-image-classifier 
        src="https://s3.us-east-2.amazonaws.com/deep-obj-cat-expt-imgs/mturk-images/${image_url}"
        categories="['Left', 'Right']"
        header="Choose which pair of shapes are more similar to each other?"
        name="category">

       <!-- Use the short-instructions section for quick instructions that the Worker
              will see while working on the task. Including some basic examples of 
              good and bad answers here can help get good results. You can include 
              any HTML here. -->
        <short-instructions>
            <p>Read the task carefully and inspect the image.</p>
            <p>Choose which pair of shapes are more similar to each other, based on overall shape.</p>
            <p>Please read full instructions before starting.</p>
        </short-instructions>

        <!-- Use the full-instructions section for more detailed instructions that the 
              Worker can open while working on the task. Including more detailed 
              instructions and additional examples of good and bad answers here can
              help get good results. You can include any HTML here. -->
        <full-instructions header="Classification Instructions">
            <p>Read the task carefully and inspect the image.</p>
            <p>There are two pairs of images, on the Left and Right -- please select the pair that,
            in your best judgement, is closest in terms of the <b>overall shape</b>.  Do not try to
            figure out what these shapes might be based on -- you should just focus on the overall
            shape and compare that with its pair.</b>.</p>
            <p>IRB information: This research will investigate how people categorize or group together different shapes.
            The results will be used to test the predictions from computer models based on how learning might work in the brain.
            It is being conducted by Dr. Randall O’Reilly of the University of Colorado, and has been reviewed and approved
            by an Institutional Review Board (“IRB”) (protocol 19-0176). You may talk to them at (303) 735-3702 or irbadmin@colorado.edu <i>if:</i>
            Your questions, concerns, or complaints are not being answered by the research team;
            You cannot reach the research team; You want to talk to someone besides the research team;
            You have questions about your rights as a research subject; You want to get information or provide input about this research.
            Any mechanical Turk worker that is over 18 years old is eligible to participate.
            During this study, you will see and respond to a series of images containing pairs of shapes,
            and you will judge which pairs of shapes are more similar to each other.
            You will receive $.01 per item or $8.00 for completing 800 items, which should take 30 minutes to complete.
            Your data will be recorded without retention of any link to your amazon ID, in a secure and encrypted format.
            There are no known or suspected risks to participants in this study beyond the risks encountered in daily living.
            Your decision to proceed with this study represents your understanding and acceptance of the above information.
            </p>
        </full-instructions>

    </crowd-image-classifier>
</crowd-form>

