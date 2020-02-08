// Handle submission of Request Demo form
  // const handleRequestDemoFormSubmit = (ev: any) => {
  //   // We don't want to let default form submission happen here, which would refresh the page.
  //   ev.preventDefault();

  //   // TODO: send email to Beneath

  //   return
  // }

  // const RequestDemoForm = (
  //   <div>
  //     <form onSubmit={handleRequestDemoFormSubmit}>
  //       <TextField
  //         id="workemail"
  //         label="Your work email"
  //         margin="normal"
  //         fullWidth
  //         required
  //       />
  //       <TextField
  //         id="phonenumber"
  //         label="Your phone number"
  //         margin="normal"
  //         fullWidth
  //         required
  //       />
  //       <TextField
  //         id="fullname"
  //         label="Your name"
  //         margin="normal"
  //         fullWidth
  //         required
  //       />
  //       <TextField
  //         id="companyname"
  //         label="Company name"
  //         margin="normal"
  //         fullWidth
  //         required
  //       />
  //       <button>Submit</button>
  //     </form>
  //     <Button
  //       variant="contained"
  //       onClick={() => {
  //         setValues({ ...values, ...{ isRequestingDemo: false } })
  //       }}>
  //       Back
  //     </Button>
  //   </div>
  // )