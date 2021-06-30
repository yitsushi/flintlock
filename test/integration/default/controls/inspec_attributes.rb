control "inspec_attributes" do
    desc "A demonstration of how InSpec attributes are mapped to Terraform outputs"
  
    describe attribute("static_terraform_output") do
      it { should eq "static terraform output" }
    end
  
    describe attribute("customized_inspec_attribute") do
      it { should eq "static terraform output" }
    end
  end